package graphql

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

const numRoutines = 5
const numRepositories = 10
const numCommits = 10 // per repo
const numPaths = 10   // per commit

func TestCachedLocationResolver(t *testing.T) {
	t.Cleanup(func() {
		db.Mocks.Repos.Get = nil
		git.Mocks.ResolveRevision = nil
		backend.Mocks.Repos.GetCommit = nil
	})

	var repoCalls uint32
	db.Mocks.Repos.Get = func(v0 context.Context, id api.RepoID) (*types.Repo, error) {
		atomic.AddUint32(&repoCalls, 1)
		return &types.Repo{ID: id}, nil
	}

	git.Mocks.ResolveRevision = func(spec string, opt *git.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID(spec), nil
	}

	var commitCalls uint32
	backend.Mocks.Repos.GetCommit = func(v0 context.Context, repo *types.Repo, commitID api.CommitID) (*git.Commit, error) {
		atomic.AddUint32(&commitCalls, 1)
		return &git.Commit{ID: commitID}, nil
	}

	cachedResolver := NewCachedLocationResolver()

	var repositoryIDs []api.RepoID
	for i := 1; i <= numRepositories; i++ {
		repositoryIDs = append(repositoryIDs, api.RepoID(i))
	}

	var commits []string
	for i := 1; i <= numCommits; i++ {
		commits = append(commits, fmt.Sprintf("%040d", i))
	}

	var paths []string
	for i := 1; i <= numPaths; i++ {
		paths = append(paths, fmt.Sprintf("/foo/%d/bar/baz.go", i))
	}

	type resolverPair struct {
		key      string
		resolver *gql.GitTreeEntryResolver
	}
	resolvers := make(chan resolverPair, numRoutines*len(repositoryIDs)*len(commits)*len(paths))

	var wg sync.WaitGroup
	errs := make(chan error, numRoutines)
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for _, repositoryID := range repositoryIDs {
				repositoryResolver, err := cachedResolver.Repository(context.Background(), repositoryID)
				if err != nil {
					errs <- err
					return
				}
				if repositoryResolver.Type().ID != repositoryID {
					errs <- fmt.Errorf("unexpected repository id. want=%d have=%d", repositoryID, repositoryResolver.Type().ID)
					return
				}

				for _, commit := range commits {
					commitResolver, err := cachedResolver.Commit(context.Background(), repositoryID, commit)
					if err != nil {
						errs <- err
						return
					}
					if commitResolver.OID() != graphqlbackend.GitObjectID(commit) {
						errs <- fmt.Errorf("unexpected commit. want=%s have=%s", commit, commitResolver.OID())
						return
					}

					for _, path := range paths {
						treeResolver, err := cachedResolver.Path(context.Background(), repositoryID, commit, path)
						if err != nil {
							errs <- err
							return
						}
						if treeResolver.Path() != path {
							errs <- fmt.Errorf("unexpected path. want=%s have=%s", path, treeResolver.Path())
							return
						}

						resolvers <- resolverPair{key: fmt.Sprintf("%d:%s:%s", repositoryID, commit, path), resolver: treeResolver}
					}
				}
			}
		}()
	}
	wg.Wait()

	close(errs)
	for err := range errs {
		t.Error(err)
	}

	if val := atomic.LoadUint32(&repoCalls); val != uint32(len(repositoryIDs)) {
		t.Errorf("unexpected number of repo calls. want=%d have=%d", len(repositoryIDs), val)
	}

	if val := atomic.LoadUint32(&commitCalls); val != uint32(len(repositoryIDs)*len(commits)) {
		t.Errorf("unexpected number of commit calls. want=%d have=%d", len(commits), val)
	}

	close(resolvers)
	resolversByKey := map[string][]*gql.GitTreeEntryResolver{}
	for pair := range resolvers {
		resolversByKey[pair.key] = append(resolversByKey[pair.key], pair.resolver)
	}

	for _, vs := range resolversByKey {
		for _, v := range vs {
			if v != vs[0] {
				t.Errorf("resolvers for same key unexpectedly have differing addresses: %p and %p", v, vs[0])
			}
		}
	}
}
