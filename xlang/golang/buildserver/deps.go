package buildserver

import (
	"context"
	"fmt"
	"go/build"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/langserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

type keyMutex struct {
	mu  sync.Mutex
	mus map[string]*sync.Mutex
}

// get returns a mutex unique to the given key.
func (k *keyMutex) get(key string) *sync.Mutex {
	k.mu.Lock()
	mu, ok := k.mus[key]
	if !ok {
		mu = &sync.Mutex{}
		k.mus[key] = mu
	}
	k.mu.Unlock()
	return mu
}

func newKeyMutex() *keyMutex {
	return &keyMutex{
		mus: map[string]*sync.Mutex{},
	}
}

type importKey struct {
	path, srcDir string
	mode         build.ImportMode
}

type importResult struct {
	pkg *build.Package
	err error
}

type depCache struct {
	importCacheMu sync.Mutex
	importCache   map[importKey]importResult

	// A mapping of package path -> direct import records
	collectReferences bool
	seenMu            sync.Mutex
	seen              map[string][]importRecord
	entryPackageDirs  []string

	// Cache for fetching Go meta tag results.
	fetchMetaCacheImportMu *keyMutex
	fetchMetaCacheMu       sync.RWMutex
	fetchMetaCache         map[string]fetchMetaResult
}

func newDepCache() *depCache {
	return &depCache{
		importCache: map[importKey]importResult{},
		seen:        map[string][]importRecord{},
		fetchMetaCacheImportMu: newKeyMutex(),
		fetchMetaCache:         map[string]fetchMetaResult{},
	}
}

// fetchTransitiveDepsOfFile fetches the transitive dependencies of
// the named Go file. A Go file's dependencies are the imports of its
// own package, plus all of its imports' imports, and so on.
//
// It adds fetched dependencies to its own file system overlay, and
// the returned depFiles should be passed onto the language server to
// add to its overlay.
func (h *BuildHandler) fetchTransitiveDepsOfFile(ctx context.Context, fileURI string, dc *depCache) (err error) {
	parentSpan := opentracing.SpanFromContext(ctx)
	span := parentSpan.Tracer().StartSpan("xlang-go: fetch transitive dependencies",
		opentracing.Tags{"fileURI": fileURI},
		opentracing.ChildOf(parentSpan.Context()),
	)
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogEvent(fmt.Sprintf("error: %v", err))
		}
		span.Finish()
	}()

	bctx := h.lang.BuildContext(ctx)
	bpkg, err := langserver.ContainingPackage(bctx, h.FilePath(fileURI))
	if err != nil && !isMultiplePackageError(err) {
		return err
	}

	err = doDeps(bpkg, 0, dc, func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		return h.doFindPackage(ctx, bctx, path, srcDir, mode, dc)
	})
	return err
}

type findPkgKey struct {
	importPath string // e.g. "github.com/gorilla/mux"
	fromDir    string // e.g. "/gopath/src/github.com/kubernetes/kubernetes"
	mode       build.ImportMode
}

type findPkgValue struct {
	ready chan struct{} // closed to broadcast readiness
	bp    *build.Package
	err   error
}

func (h *BuildHandler) findPackageCached(ctx context.Context, bctx *build.Context, p, srcDir string, mode build.ImportMode) (*build.Package, error) {
	// bctx.FindPackage and loader.Conf does not have caching, and due to
	// vendor we need to repeat work. So what we do is normalise the
	// srcDir w.r.t. potential vendoring. This makes the assumption that
	// the underlying FS for bctx is always the same, which is currently a
	// correct assumption. That may have to be revisited with zap.
	//
	// Example: A project gh.com/p/r has a single vendor folder at
	// /goroot/gh.com/p/r/vendor. Both gh.com/p/r/foo and
	// gh.com/p/r/bar/baz use gh.com/gorilla/mux.
	// loader will then call both
	//
	//   findPackage(..., "gh.com/gorilla/mux", "/gopath/src/gh.com/p/r/foo", ...)
	//   findPackage(..., "gh.com/gorilla/mux", "/gopath/src/gh.com/p/r/bar/baz", ...)
	//
	// findPackage then starts from the directory and checks for any
	// potential vendor directories which contains
	// "gh.com/gorilla/mux". Given "/gopath/src/gh.com/p/r/foo" and
	// "/gopath/src/gh.com/p/r/bar/baz" may have different vendor dirs to
	// check, it can't cache this work.
	//
	// So instead of passing "/gopath/src/gh.com/p/r/bar/baz" we pass
	// "/gopath/src/gh.com/p/r" because we know the first vendor dir to
	// check is "/gopath/src/gh.com/p/r/vendor". This also means that
	// "/gopath/src/gh.com/p/r/bar/baz" and "/gopath/src/gh.com/p/r/foo"
	// get the same cache key findPkgKey{"gh.com/gorilla/mux", "/gopath/src/gh.com/p/r", 0}.
	if !build.IsLocalImport(p) {
		gopathSrc := path.Join(gopath, "src")
		for !bctx.IsDir(path.Join(srcDir, "vendor", p)) && srcDir != gopathSrc && srcDir != goroot && srcDir != "/" {
			srcDir = path.Dir(srcDir)
		}
	}

	// We do single-flighting as well. conf.Loader does the same, but its
	// single-flighting is based on srcDir before it is normalised.
	k := findPkgKey{p, srcDir, mode}
	h.findPkgMu.Lock()
	if h.findPkg == nil {
		h.findPkg = make(map[findPkgKey]*findPkgValue)
	}
	v, ok := h.findPkg[k]
	if ok {
		h.findPkgMu.Unlock()
		<-v.ready

		return v.bp, v.err
	}

	v = &findPkgValue{ready: make(chan struct{})}
	h.findPkg[k] = v
	h.findPkgMu.Unlock()

	v.bp, v.err = h.findPackage(ctx, bctx, p, srcDir, mode)

	close(v.ready)
	return v.bp, v.err
}

// findPackage is a langserver.FindPackageFunc which integrates with the build
// server. It will fetch dependencies just in time.
func (h *BuildHandler) findPackage(ctx context.Context, bctx *build.Context, path, srcDir string, mode build.ImportMode) (*build.Package, error) {
	return h.doFindPackage(ctx, bctx, path, srcDir, mode, newDepCache())
}

func (h *BuildHandler) doFindPackage(ctx context.Context, bctx *build.Context, path, srcDir string, mode build.ImportMode, dc *depCache) (*build.Package, error) {
	// If the package exists in the repo, or is vendored, or has
	// already been fetched, this will succeed.
	pkg, err := bctx.Import(path, srcDir, mode)
	if isMultiplePackageError(err) {
		err = nil
	}
	if err == nil {
		return pkg, nil
	}

	// If this package resolves to the same repo, then use any
	// imported package, even if it has errors. The errors would
	// be caused by the repo itself, not our dep fetching.
	//
	// TODO(sqs): if a package example.com/a imports
	// example.com/a/b and example.com/a/b lives in a separate
	// repo, then this will break. This is the case for some
	// azul3d packages, but it's rare.
	if langserver.PathHasPrefix(path, h.rootImportPath) {
		if pkg != nil {
			return pkg, nil
		}
		return nil, fmt.Errorf("package %q is inside of workspace root but failed to import: %s", path, err)
	}

	// Otherwise, it's an external dependency. Fetch the package
	// and try again.
	d, err := resolveImportPath(http.DefaultClient, path, dc)
	if err != nil {
		return nil, err
	}

	// If this package resolves to the same repo, then don't fetch
	// it; it is already on disk. If we fetch it, we might end up
	// with multiple conflicting versions of the workspace's repo
	// overlaid on each other.
	if langserver.PathHasPrefix(d.projectRoot, h.rootImportPath) {
		return nil, fmt.Errorf("package %q is inside of workspace root, refusing to fetch remotely", path)
	}

	urlMu := h.depURLMutex.get(d.cloneURL)
	urlMu.Lock()
	defer urlMu.Unlock()

	// Check again after waiting.
	pkg, err = bctx.Import(path, srcDir, mode)
	if err == nil {
		return pkg, nil
	}

	// We may have a specific rev to use (from glide.lock)
	if rev := h.pinnedDep(ctx, d.importPath); rev != "" {
		d.rev = rev
	}

	// If not, we hold the lock and we will fetch the dep.
	if err := h.fetchDep(ctx, d); err != nil {
		return nil, err
	}

	pkg, err = bctx.Import(path, srcDir, mode)
	if isMultiplePackageError(err) {
		err = nil
	}
	return pkg, err
}

func (h *BuildHandler) fetchDep(ctx context.Context, d *directory) error {
	if d.vcs != "git" {
		return fmt.Errorf("dependency at import path %q has unsupported VCS %q (clone URL is %q)", d.importPath, d.vcs, d.cloneURL)
	}

	rev := d.rev
	if rev == "" {
		rev = "HEAD"
	}
	cloneURL, err := url.Parse(d.cloneURL)
	if err != nil {
		return err
	}
	fs, err := NewDepRepoVFS(cloneURL, rev)
	if err != nil {
		return err
	}

	if _, isStdlib := stdlibPackagePaths[d.importPath]; isStdlib {
		fs = addSysZversionFile(fs)
	}

	var oldPath string
	_, isStdlib := stdlibPackagePaths[d.importPath]
	if isStdlib {
		oldPath = goroot // stdlib
	} else {
		oldPath = path.Join(gopath, "src", d.projectRoot) // non-stdlib
	}

	h.HandlerShared.Mu.Lock()
	h.FS.Bind(oldPath, fs, "/", ctxvfs.BindAfter)
	if !isStdlib {
		h.gopathDeps = append(h.gopathDeps, d)
	}
	h.HandlerShared.Mu.Unlock()

	return nil
}

func (h *BuildHandler) pinnedDep(ctx context.Context, pkg string) string {
	h.pinnedDepsOnce.Do(func() {
		h.HandlerShared.Mu.Lock()
		fs := h.FS
		root := h.RootFSPath
		h.HandlerShared.Mu.Unlock()

		// We assume glide.lock is in the top-level dir of the
		// repo. This assumption may not be valid in the future.
		yml, err := ctxvfs.ReadFile(ctx, fs, path.Join(root, "glide.lock"))
		if err == nil && len(yml) > 0 {
			h.pinnedDeps = loadGlideLock(yml)
			return
		}

		// Next we try load from Godeps. Note: We will mount the wrong
		// dependencies in these two strange cases:
		// 1. Different revisions for pkgs in the same repo.
		// 2. Using a pkg not in Godeps, but another pkg from the same repo is in Godeps
		// In both cases, we use the revision for the pkg we first try and fetch.
		b, err := ctxvfs.ReadFile(ctx, fs, path.Join(root, "Godeps/Godeps.json"))
		if err == nil && len(b) > 0 {
			h.pinnedDeps = loadGodeps(b)
			return
		}
	})
	return h.pinnedDeps.Find(pkg)
}

func doDeps(pkg *build.Package, mode build.ImportMode, dc *depCache, importPackage func(path, srcDir string, mode build.ImportMode) (*build.Package, error)) error {
	// Separate mutexes for each package import path.
	importPathMutex := newKeyMutex()

	gate := make(chan struct{}, runtime.GOMAXPROCS(0)) // I/O concurrency limit
	cachedImportPackage := func(path, srcDir string, mode build.ImportMode) (*build.Package, error) {
		dc.importCacheMu.Lock()
		res, ok := dc.importCache[importKey{path, srcDir, mode}]
		dc.importCacheMu.Unlock()
		if ok {
			return res.pkg, res.err
		}

		gate <- struct{}{} // limit I/O concurrency
		defer func() { <-gate }()

		mu := importPathMutex.get(path)
		mu.Lock() // only try to import a path once
		defer mu.Unlock()

		dc.importCacheMu.Lock()
		res, ok = dc.importCache[importKey{path, srcDir, mode}]
		dc.importCacheMu.Unlock()
		if !ok {
			res.pkg, res.err = importPackage(path, srcDir, mode)
			dc.importCacheMu.Lock()
			dc.importCache[importKey{path, srcDir, mode}] = res
			dc.importCacheMu.Unlock()
		}
		return res.pkg, res.err
	}

	var errs errorList
	var wg sync.WaitGroup
	var do func(pkg *build.Package)
	do = func(pkg *build.Package) {
		dc.seenMu.Lock()
		if _, seen := dc.seen[pkg.Dir]; seen {
			dc.seenMu.Unlock()
			return
		}
		dc.seen[pkg.Dir] = []importRecord{}
		dc.seenMu.Unlock()

		for _, path := range allPackageImportsSorted(pkg) {
			if path == "C" {
				continue
			}
			wg.Add(1)
			parentPkg := pkg
			go func(path string) {
				defer wg.Done()
				pkg, err := cachedImportPackage(path, pkg.Dir, mode)
				if err != nil {
					errs.add(err)
				}
				if pkg != nil {
					if dc.collectReferences {
						dc.seenMu.Lock()
						dc.seen[parentPkg.Dir] = append(dc.seen[parentPkg.Dir], importRecord{pkg: parentPkg, imports: pkg})
						dc.seenMu.Unlock()
					}
					do(pkg)
				}
			}(path)
		}
	}
	do(pkg)
	if dc.collectReferences {
		dc.seenMu.Lock()
		dc.entryPackageDirs = append(dc.entryPackageDirs, pkg.Dir)
		dc.seenMu.Unlock()
	}
	wg.Wait()
	return errs.error()
}

func allPackageImportsSorted(pkg *build.Package) []string {
	uniq := map[string]struct{}{}
	for _, p := range pkg.Imports {
		uniq[p] = struct{}{}
	}
	for _, p := range pkg.TestImports {
		uniq[p] = struct{}{}
	}
	for _, p := range pkg.XTestImports {
		uniq[p] = struct{}{}
	}
	imps := make([]string, 0, len(uniq))
	for p := range uniq {
		imps = append(imps, p)
	}
	sort.Strings(imps)
	return imps
}

func isMultiplePackageError(err error) bool {
	_, ok := err.(*build.MultiplePackageError)
	return ok
}

// FetchCommonDeps will fetch our common used dependencies. This is to avoid
// impacting the first ever typecheck we do in a repo since it will have to
// fetch the dependency from the internet.
func FetchCommonDeps() {
	// github.com/golang/go
	d, _ := resolveStaticImportPath("time")
	u, _ := url.Parse(d.cloneURL)
	_, _ = NewDepRepoVFS(u, d.rev)
}

// NewDepRepoVFS returns a virtual file system interface for accessing
// the files in the specified (public) repo at the given commit.
//
// TODO(sqs): design a way for the Go build/lang server to access
// private repos. Private repos are currently only supported for the
// main workspace repo, not as dependencies.
var NewDepRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
	// Fast-path for GitHub repos, which we can fetch on-demand from
	// GitHub's repo .zip archive download endpoint.
	if cloneURL.Host == "github.com" {
		fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
		return vfsutil.NewGitHubRepoVFS(fullName, rev, "", true)
	}

	// Fall back to a full git clone for non-github.com repos.
	return &vfsutil.GitRepoVFS{
		CloneURL: cloneURL.String(),
		Rev:      rev,
	}, nil
}
