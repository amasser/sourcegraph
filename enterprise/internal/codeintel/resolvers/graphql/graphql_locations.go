package resolvers

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

func resolveLocations(ctx context.Context, collectionResolver *repositoryCollectionResolver, locations []resolvers.AdjustedLocation) ([]gql.LocationResolver, error) {
	resolvedLocations := make([]gql.LocationResolver, 0, len(locations))
	for i := range locations {
		resolver, err := resolveLocation(ctx, collectionResolver, locations[i])
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

func resolveLocation(ctx context.Context, collectionResolver *repositoryCollectionResolver, location resolvers.AdjustedLocation) (gql.LocationResolver, error) {
	treeResolver, err := collectionResolver.resolve(ctx, api.RepoID(location.Dump.RepositoryID), location.AdjustedCommit, location.Path)
	if err != nil {
		return nil, err
	}
	if treeResolver == nil {
		return nil, nil
	}

	return gql.NewLocationResolver(treeResolver, &location.AdjustedRange), nil
}

type repositoryCollectionResolver struct {
	m                         sync.RWMutex
	commitCollectionResolvers map[api.RepoID]*commitCollectionResolver
}

func NewRepositoryCollectionResolver() *repositoryCollectionResolver {
	return &repositoryCollectionResolver{
		commitCollectionResolvers: map[api.RepoID]*commitCollectionResolver{},
	}
}

// resolve returns a GitTreeEntryResolver for the given repository, commit, and path. This will cache
// the repository, commit, and path resolvers if they have been previously constructed with this same
// struct instance. If the commit resolver cannot be constructed, a nil resolver is returned.
func (r *repositoryCollectionResolver) resolve(ctx context.Context, repoID api.RepoID, commit, path string) (*gql.GitTreeEntryResolver, error) {
	commitCollectionResolver, err := r.resolveRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	pathCollectionResolver, err := commitCollectionResolver.resolveCommit(ctx, commit)
	if err != nil {
		return nil, err
	}

	return pathCollectionResolver.resolvePath(ctx, path)
}

// resolveRepository returns a commitCollectionResolver with the given resolved repository.
func (r *repositoryCollectionResolver) resolveRepository(ctx context.Context, repoID api.RepoID) (*commitCollectionResolver, error) {
	r.m.RLock()
	if payload, ok := r.commitCollectionResolvers[repoID]; ok {
		r.m.RUnlock()
		return payload, nil
	}
	r.m.RUnlock()

	r.m.Lock()
	defer r.m.Unlock()
	if payload, ok := r.commitCollectionResolvers[repoID]; ok {
		return payload, nil
	}

	repositoryResolver, err := resolveRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	payload := &commitCollectionResolver{
		repositoryResolver:      repositoryResolver,
		pathCollectionResolvers: map[string]*pathCollectionResolver{},
	}

	r.commitCollectionResolvers[repoID] = payload
	return payload, nil
}

type commitCollectionResolver struct {
	repositoryResolver *gql.RepositoryResolver

	m                       sync.RWMutex
	pathCollectionResolvers map[string]*pathCollectionResolver
}

// resolveCommit returns a pathCollectionResolver with the given resolved commit.
func (r *commitCollectionResolver) resolveCommit(ctx context.Context, commit string) (*pathCollectionResolver, error) {
	r.m.RLock()
	if resolver, ok := r.pathCollectionResolvers[commit]; ok {
		r.m.RUnlock()
		return resolver, nil
	}
	r.m.RUnlock()

	r.m.Lock()
	defer r.m.Unlock()
	if resolver, ok := r.pathCollectionResolvers[commit]; ok {
		return resolver, nil
	}

	commitResolver, err := resolveCommitFrom(ctx, r.repositoryResolver, commit)
	if err != nil {
		return nil, err
	}

	resolver := &pathCollectionResolver{
		commitResolver: commitResolver,
		pathResolvers:  map[string]*gql.GitTreeEntryResolver{},
	}

	r.pathCollectionResolvers[commit] = resolver
	return resolver, nil
}

type pathCollectionResolver struct {
	commitResolver *gql.GitCommitResolver

	m             sync.RWMutex
	pathResolvers map[string]*gql.GitTreeEntryResolver
}

// pathCollectionResolver returns a GitTreeEntryResolver with the given path. If the
// commit resolver could not be constructed, a nil resolver is returned.
func (r *pathCollectionResolver) resolvePath(ctx context.Context, path string) (*gql.GitTreeEntryResolver, error) {
	r.m.RLock()
	if resolver, ok := r.pathResolvers[path]; ok {
		r.m.RUnlock()
		return resolver, nil
	}
	r.m.RUnlock()

	r.m.Lock()
	defer r.m.Unlock()
	if resolver, ok := r.pathResolvers[path]; ok {
		return resolver, nil
	}

	resolver, err := resolvePathFrom(ctx, r.commitResolver, path)
	if err != nil {
		return nil, err
	}

	r.pathResolvers[path] = resolver
	return resolver, nil
}

// resolveRepository returns a repository resolver for the given name.
func resolveRepository(ctx context.Context, repoID api.RepoID) (*gql.RepositoryResolver, error) {
	repo, err := backend.Repos.Get(ctx, repoID)
	if err != nil {
		return nil, err
	}

	return gql.NewRepositoryResolver(repo), nil
}

// resolveCommit returns the GitCommitResolver for the given repository and commit. If the
// commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func resolveCommit(ctx context.Context, repositoryResolver *gql.RepositoryResolver, commit string) (*gql.GitCommitResolver, error) {
	return resolveCommitFrom(ctx, repositoryResolver, commit)
}

// resolveCommitFrom returns the GitCommitResolver for the given repository resolver and commit.
// If the commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func resolveCommitFrom(ctx context.Context, repositoryResolver *gql.RepositoryResolver, commit string) (*gql.GitCommitResolver, error) {
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

// resolvePath returns the GitTreeResolver for the given repository, commit, and path. If the
// commit does not exist for the repository, a nil resolver is returned. Any other error is
// returned unmodified.
func resolvePath(ctx context.Context, repoID api.RepoID, commit, path string) (*gql.GitTreeEntryResolver, error) {
	repositoryResolver, err := resolveRepository(ctx, repoID)
	if err != nil {
		return nil, err
	}

	commitResolver, err := resolveCommit(ctx, repositoryResolver, commit)
	if err != nil {
		return nil, err
	}

	return resolvePathFrom(ctx, commitResolver, path)
}

// resolvePath returns the GitTreeResolver for the given commit resolver, and path. If the
// commit resolver is nil, a nil resolver is returned. Any other error is returned unmodified.
func resolvePathFrom(ctx context.Context, commitResolver *gql.GitCommitResolver, path string) (*gql.GitTreeEntryResolver, error) {
	if commitResolver == nil {
		return nil, nil
	}

	return gql.NewGitTreeEntryResolver(commitResolver, gql.CreateFileInfo(path, true)), nil
}
