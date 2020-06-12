package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type gqlUploadConnectionResolver struct{ resolver *uploadsResolver }

func NewGraphQLUploadConnectionResolver(resolver *uploadsResolver) gql.LSIFUploadConnectionResolver {
	return &gqlUploadConnectionResolver{resolver: resolver}
}

func (r *gqlUploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.resolver.uploads))
	for i := range r.resolver.uploads {
		resolvers = append(resolvers, NewGraphQLUploadResolver(r.resolver.uploads[i]))
	}

	return resolvers, nil
}

func (r *gqlUploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.totalCount), nil
}

func (r *gqlUploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(r.resolver.nextOffset), nil
}

//
//

type gqlIndexConnectionResolver struct{ resolver *indexesResolver }

func NewGraphQLIndexConnectionResolver(resolver *indexesResolver) gql.LSIFIndexConnectionResolver {
	return &gqlIndexConnectionResolver{resolver: resolver}
}

func (r *gqlIndexConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFIndexResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFIndexResolver, 0, len(r.resolver.indexes))
	for i := range r.resolver.indexes {
		resolvers = append(resolvers, NewGraphQLIndexResolver(r.resolver.indexes[i]))
	}

	return resolvers, nil
}

func (r *gqlIndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.totalCount), nil
}

func (r *gqlIndexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(r.resolver.nextOffset), nil
}

//
//

type gqlLocationConnectionResolver struct {
	locations []AdjustedLocation
	cursor    string
}

func NewGraphQLLocationConnectionResolver(locations []AdjustedLocation, cursor string) gql.LocationConnectionResolver {
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
	diagnostics []AdjustedDiagnostic
	totalCount  int
}

func NewGraphQLDiagnosticConnectionResolver(diagnostics []AdjustedDiagnostic, totalCount int) gql.DiagnosticConnectionResolver {
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
