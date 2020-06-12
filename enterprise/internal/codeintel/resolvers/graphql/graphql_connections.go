package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type UploadConnectionResolver struct{ resolver *resolvers.UploadsResolver }

func NewUploadConnectionResolver(resolver *resolvers.UploadsResolver) gql.LSIFUploadConnectionResolver {
	return &UploadConnectionResolver{resolver: resolver}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.resolver.Uploads))
	for i := range r.resolver.Uploads {
		resolvers = append(resolvers, NewUploadResolver(r.resolver.Uploads[i]))
	}

	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(r.resolver.NextOffset), nil
}

//
//

type IndexConnectionResolver struct{ resolver *resolvers.IndexesResolver }

func NewIndexConnectionResolver(resolver *resolvers.IndexesResolver) gql.LSIFIndexConnectionResolver {
	return &IndexConnectionResolver{resolver: resolver}
}

func (r *IndexConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFIndexResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFIndexResolver, 0, len(r.resolver.Indexes))
	for i := range r.resolver.Indexes {
		resolvers = append(resolvers, NewIndexResolver(r.resolver.Indexes[i]))
	}

	return resolvers, nil
}

func (r *IndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.TotalCount), nil
}

func (r *IndexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(r.resolver.NextOffset), nil
}

//
//

type LocationConnectionResolver struct {
	locations []resolvers.AdjustedLocation
	cursor    *string
}

func NewLocationConnectionResolver(locations []resolvers.AdjustedLocation, cursor *string) gql.LocationConnectionResolver {
	return &LocationConnectionResolver{locations: locations, cursor: cursor}
}

func (r *LocationConnectionResolver) Nodes(ctx context.Context) ([]gql.LocationResolver, error) {
	return resolveLocations(ctx, NewRepositoryCollectionResolver(), r.locations)
}

func (r *LocationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return encodeCursor(r.cursor), nil
}

//
//

type DiagnosticConnectionResolver struct {
	diagnostics []resolvers.AdjustedDiagnostic
	totalCount  int
}

func NewDiagnosticConnectionResolver(diagnostics []resolvers.AdjustedDiagnostic, totalCount int) gql.DiagnosticConnectionResolver {
	return &DiagnosticConnectionResolver{diagnostics: diagnostics, totalCount: totalCount}
}

func (r *DiagnosticConnectionResolver) Nodes(ctx context.Context) ([]gql.DiagnosticResolver, error) {
	collectionResolver := NewRepositoryCollectionResolver()

	resolvers := make([]gql.DiagnosticResolver, 0, len(r.diagnostics))
	for i := range r.diagnostics {
		resolvers = append(resolvers, NewDiagnosticResolver(r.diagnostics[i], collectionResolver))
	}

	return resolvers, nil
}

func (r *DiagnosticConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.totalCount), nil
}

func (r *DiagnosticConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(len(r.diagnostics) < r.totalCount), nil
}
