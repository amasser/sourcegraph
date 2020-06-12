package graphql

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// TODO - document
type CachedLocationResolver struct {
	sync.RWMutex
	children map[api.RepoID]*cachedRepositoryResolver
}

type cachedRepositoryResolver struct {
	sync.RWMutex
	resolver *gql.RepositoryResolver
	children map[string]*cachedCommitResolver
}

type cachedCommitResolver struct {
	sync.RWMutex
	resolver *gql.GitCommitResolver
	children map[string]*gql.GitTreeEntryResolver
}

func NewCachedLocationResolver() *CachedLocationResolver {
	return &CachedLocationResolver{
		children: map[api.RepoID]*cachedRepositoryResolver{},
	}
}

// TODO - test
// TODO - document
func (r *CachedLocationResolver) Repository(ctx context.Context, id api.RepoID) (*gql.RepositoryResolver, error) {
	cachedRepositoryResolver, err := r.cachedRepository(ctx, id)
	if err != nil {
		return nil, err
	}
	return cachedRepositoryResolver.resolver, nil
}

// TODO - document
func (r *CachedLocationResolver) Commit(ctx context.Context, id api.RepoID, commit string) (*gql.GitCommitResolver, error) {
	cachedCommitResolver, err := r.cachedCommit(ctx, id, commit)
	if err != nil {
		return nil, err
	}
	return cachedCommitResolver.resolver, nil
}

// TODO - document
func (r *CachedLocationResolver) Path(ctx context.Context, id api.RepoID, commit, path string) (*gql.GitTreeEntryResolver, error) {
	pathResolver, err := r.cachedPath(ctx, id, commit, path)
	if err != nil {
		return nil, err
	}
	return pathResolver, nil
}

// TODO - document
func (r *CachedLocationResolver) cachedRepository(ctx context.Context, id api.RepoID) (*cachedRepositoryResolver, error) {
	// Fast-path cache check
	r.RLock()
	cr, ok := r.children[id]
	r.RUnlock()
	if ok {
		return cr, nil
	}

	r.Lock()
	defer r.Unlock()

	// Check again once locked to avoid race
	if resolver, ok := r.children[id]; ok {
		return resolver, nil
	}

	// Resolve new value and store in cache
	resolver, err := r.resolveRepository(ctx, id)
	if err != nil {
		return nil, err
	}
	cachedResolver := &cachedRepositoryResolver{resolver: resolver, children: map[string]*cachedCommitResolver{}}
	r.children[id] = cachedResolver
	return cachedResolver, nil
}

// TODO - document
func (r *CachedLocationResolver) cachedCommit(ctx context.Context, id api.RepoID, commit string) (*cachedCommitResolver, error) {
	parentResolver, err := r.cachedRepository(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fast-path cache check
	parentResolver.RLock()
	cr, ok := parentResolver.children[commit]
	parentResolver.RUnlock()
	if ok {
		return cr, nil
	}

	parentResolver.Lock()
	defer parentResolver.Unlock()

	// Check again once locked to avoid race
	if resolver, ok := parentResolver.children[commit]; ok {
		return resolver, nil
	}

	// Resolve new value and store in cache
	resolver, err := r.resolveCommit(ctx, parentResolver.resolver, commit)
	if err != nil {
		return nil, err
	}
	cachedResolver := &cachedCommitResolver{resolver: resolver, children: map[string]*gql.GitTreeEntryResolver{}}
	parentResolver.children[commit] = cachedResolver
	return cachedResolver, nil
}

// TODO - document
func (r *CachedLocationResolver) cachedPath(ctx context.Context, id api.RepoID, commit, path string) (*gql.GitTreeEntryResolver, error) {
	parentResolver, err := r.cachedCommit(ctx, id, commit)
	if err != nil || parentResolver == nil {
		return nil, err
	}

	// Fast-path cache check
	parentResolver.Lock()
	cr, ok := parentResolver.children[path]
	parentResolver.Unlock()
	if ok {
		return cr, nil
	}

	parentResolver.Lock()
	defer parentResolver.Unlock()

	// Check again once locked to avoid race
	if resolver, ok := parentResolver.children[path]; ok {
		return resolver, nil
	}

	// Resolve new value and store in cache
	resolver, err := r.resolvePath(ctx, parentResolver.resolver, path)
	if err != nil {
		return nil, err
	}
	parentResolver.children[path] = resolver
	return resolver, nil
}

// TODO - document
func (r *CachedLocationResolver) resolveRepository(ctx context.Context, id api.RepoID) (*gql.RepositoryResolver, error) {
	repo, err := backend.Repos.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return gql.NewRepositoryResolver(repo), nil
}

// TODO - document
func (r *CachedLocationResolver) resolveCommit(ctx context.Context, repositoryResolver *gql.RepositoryResolver, commit string) (*gql.GitCommitResolver, error) {
	gitserverRepo, err := backend.CachedGitRepo(ctx, repositoryResolver.Type())
	if err != nil {
		return nil, err
	}

	commitID, err := git.ResolveRevision(ctx, *gitserverRepo, nil, commit, &git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		if gitserver.IsRevisionNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return repositoryResolver.CommitFromID(ctx, &gql.RepositoryCommitArgs{Rev: commit}, commitID)
}

// TODO - document
func (r *CachedLocationResolver) resolvePath(ctx context.Context, commitResolver *gql.GitCommitResolver, path string) (*gql.GitTreeEntryResolver, error) {
	return gql.NewGitTreeEntryResolver(commitResolver, gql.CreateFileInfo(path, true)), nil
}

// TODO - test
// TODO - document
func resolveLocation(ctx context.Context, locationResolver *CachedLocationResolver, location resolvers.AdjustedLocation) (gql.LocationResolver, error) {
	treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.Dump.RepositoryID), location.AdjustedCommit, location.Path)
	if err != nil || treeResolver == nil {
		return nil, err
	}

	return gql.NewLocationResolver(treeResolver, &location.AdjustedRange), nil
}

// TODO - test
// TODO - document
func resolveLocations(ctx context.Context, locationResolver *CachedLocationResolver, locations []resolvers.AdjustedLocation) ([]gql.LocationResolver, error) {
	resolvedLocations := make([]gql.LocationResolver, 0, len(locations))
	for i := range locations {
		resolver, err := resolveLocation(ctx, locationResolver, locations[i])
		if err != nil {
			return nil, err
		}
		if resolver == nil {
			continue
		}

		resolvedLocations = append(resolvedLocations, resolver)
	}

	return resolvedLocations, nil
}
