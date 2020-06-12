package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type gqlUploadConnectionResolver struct{ resolver *resolvers.UploadsResolver }

func NewGraphQLUploadConnectionResolver(resolver *resolvers.UploadsResolver) gql.LSIFUploadConnectionResolver {
	return &gqlUploadConnectionResolver{resolver: resolver}
}

func (r *gqlUploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.resolver.Uploads))
	for i := range r.resolver.Uploads {
		resolvers = append(resolvers, NewGraphQLUploadResolver(r.resolver.Uploads[i]))
	}

	return resolvers, nil
}

func (r *gqlUploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.TotalCount), nil
}

func (r *gqlUploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(r.resolver.NextOffset), nil
}

//
//

type gqlIndexConnectionResolver struct{ resolver *resolvers.IndexesResolver }

func NewGraphQLIndexConnectionResolver(resolver *resolvers.IndexesResolver) gql.LSIFIndexConnectionResolver {
	return &gqlIndexConnectionResolver{resolver: resolver}
}

func (r *gqlIndexConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFIndexResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFIndexResolver, 0, len(r.resolver.Indexes))
	for i := range r.resolver.Indexes {
		resolvers = append(resolvers, NewGraphQLIndexResolver(r.resolver.Indexes[i]))
	}

	return resolvers, nil
}

func (r *gqlIndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.TotalCount), nil
}

func (r *gqlIndexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(r.resolver.NextOffset), nil
}

//
//

type gqlLocationConnectionResolver struct {
	locations []resolvers.AdjustedLocation
	cursor    *string
}

func NewGraphQLLocationConnectionResolver(locations []resolvers.AdjustedLocation, cursor *string) gql.LocationConnectionResolver {
	return &gqlLocationConnectionResolver{locations: locations, cursor: cursor}
}

func (r *gqlLocationConnectionResolver) Nodes(ctx context.Context) ([]gql.LocationResolver, error) {
	return resolveLocations(ctx, NewRepositoryCollectionResolver(), r.locations)
}

func (r *gqlLocationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return encodeCursor(r.cursor), nil
}

//
//

type gqlDiagnosticConnectionResolver struct {
	diagnostics []resolvers.AdjustedDiagnostic
	totalCount  int
}

func NewGraphQLDiagnosticConnectionResolver(diagnostics []resolvers.AdjustedDiagnostic, totalCount int) gql.DiagnosticConnectionResolver {
	return &gqlDiagnosticConnectionResolver{diagnostics: diagnostics, totalCount: totalCount}
}

func (r *gqlDiagnosticConnectionResolver) Nodes(ctx context.Context) ([]gql.DiagnosticResolver, error) {
	collectionResolver := NewRepositoryCollectionResolver()

	resolvers := make([]gql.DiagnosticResolver, 0, len(r.diagnostics))
	for i := range r.diagnostics {
		resolvers = append(resolvers, NewGraphQLDiagnosticResolver(r.diagnostics[i], collectionResolver))
	}

	return resolvers, nil
}

func (r *gqlDiagnosticConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *gqlDiagnosticConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.diagnostics) < r.totalCount), nil
}
