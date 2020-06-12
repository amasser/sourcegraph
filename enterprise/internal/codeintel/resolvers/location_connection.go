package resolvers

import (
	"context"

	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type AdjustedLocation struct {
	location       codeintelapi.ResolvedLocation
	adjustedCommit string
	adjustedRange  lsp.Range
}

type locationConnectionResolver struct {
	locations []AdjustedLocation
	endCursor string
}

var _ gql.LocationConnectionResolver = &locationConnectionResolver{}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]gql.LocationResolver, error) {
	return resolveLocations(ctx, r.locations)
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if r.endCursor != "" {
		return graphqlutil.NextPageCursor(r.endCursor), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func resolveLocations(ctx context.Context, locations []AdjustedLocation) ([]gql.LocationResolver, error) {
	collectionResolver := &repositoryCollectionResolver{
		commitCollectionResolvers: map[api.RepoID]*commitCollectionResolver{},
	}

	var resovledLocations []gql.LocationResolver
	for _, location := range locations {
		treeResolver, err := collectionResolver.resolve(ctx, api.RepoID(location.location.Dump.RepositoryID), location.adjustedCommit, location.location.Path)
		if err != nil {
			return nil, err
		}

		if treeResolver == nil {
			continue
		}

		ar := location.adjustedRange
		resovledLocations = append(resovledLocations, gql.NewLocationResolver(treeResolver, &ar))
	}

	return resovledLocations, nil
}
