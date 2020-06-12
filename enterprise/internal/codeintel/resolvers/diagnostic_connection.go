package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type diagnosticConnectionResolver struct {
	totalCount  int
	diagnostics []AdjustedDiagnostic
}

var _ gql.DiagnosticConnectionResolver = &diagnosticConnectionResolver{}

func (r *diagnosticConnectionResolver) Nodes(ctx context.Context) ([]gql.DiagnosticResolver, error) {
	return resolveDiagnostics(r.diagnostics), nil
}

func (r *diagnosticConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *diagnosticConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.diagnostics) < r.totalCount), nil
}

func resolveDiagnostics(diagnostics []AdjustedDiagnostic) []gql.DiagnosticResolver {
	collectionResolver := &repositoryCollectionResolver{
		commitCollectionResolvers: map[api.RepoID]*commitCollectionResolver{},
	}

	var resolvers []gql.DiagnosticResolver
	for _, diagnostic := range diagnostics {
		resolvers = append(resolvers, &diagnosticResolver{
			diagnostic:         diagnostic,
			collectionResolver: collectionResolver,
		})
	}

	return resolvers
}
