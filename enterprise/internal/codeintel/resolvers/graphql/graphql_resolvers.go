package resolvers

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

const DefaultReferencesPageSize = 100
const DefaultDiagnosticsPageSize = 100

var ErrIllegalLimit = errors.New("illegal limit")

type gqlResolver struct {
	resolver *resolvers.Resolver
}

func NewGraphQLResolver(store store.Store, bundleManagerClient bundles.BundleManagerClient, codeIntelAPI codeintelapi.CodeIntelAPI) gql.CodeIntelResolver {
	return &gqlResolver{resolver: resolvers.NewResolver(store, bundleManagerClient, codeIntelAPI)}
}

func (r *gqlResolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (gql.LSIFUploadResolver, error) {
	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	upload, exists, err := r.resolver.GetUploadByID(ctx, int(uploadID))
	if err != nil || !exists {
		return nil, err
	}

	return NewGraphQLUploadResolver(upload), nil
}

func (r *gqlResolver) LSIFUploads(ctx context.Context, args *gql.LSIFUploadsQueryArgs) (gql.LSIFUploadConnectionResolver, error) {
	return r.LSIFUploadsByRepo(ctx, &gql.LSIFRepositoryUploadsQueryArgs{LSIFUploadsQueryArgs: args})
}

func (r *gqlResolver) LSIFUploadsByRepo(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (gql.LSIFUploadConnectionResolver, error) {
	opts, err := makeGetUploadsOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	return NewGraphQLUploadConnectionResolver(r.resolver.UploadConnectionResolver(opts)), nil
}

func (r *gqlResolver) DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may delete LSIF data for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.DeleteUploadByID(ctx, int(uploadID)); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *gqlResolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (gql.LSIFIndexResolver, error) {
	indexID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	index, exists, err := r.resolver.GetIndexByID(ctx, int(indexID))
	if err != nil || !exists {
		return nil, err
	}

	return NewGraphQLIndexResolver(index), nil
}

func (r *gqlResolver) LSIFIndexes(ctx context.Context, args *gql.LSIFIndexesQueryArgs) (gql.LSIFIndexConnectionResolver, error) {
	return r.LSIFIndexesByRepo(ctx, &gql.LSIFRepositoryIndexesQueryArgs{LSIFIndexesQueryArgs: args})
}

func (r *gqlResolver) LSIFIndexesByRepo(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (gql.LSIFIndexConnectionResolver, error) {
	opts, err := makeGetIndexesOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	return NewGraphQLIndexConnectionResolver(r.resolver.IndexConnectionResolver(opts)), nil
}

func (r *gqlResolver) DeleteLSIFIndex(ctx context.Context, id graphql.ID) (*gql.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may delete LSIF data for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	indexID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	if err := r.resolver.DeleteIndexByID(ctx, int(indexID)); err != nil {
		return nil, err
	}

	return &gql.EmptyResponse{}, nil
}

func (r *gqlResolver) GitBlobLSIFData(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (gql.GitBlobLSIFDataResolver, error) {
	resolver, err := r.resolver.QueryResolver(ctx, args)
	if err != nil || resolver == nil {
		return nil, err
	}

	return NewGraphQLQueryResolver(resolver), nil
}

//
//

type gqlQueryResolver struct {
	resolver *resolvers.QueryResolver
}

func NewGraphQLQueryResolver(resolver *resolvers.QueryResolver) gql.GitBlobLSIFDataResolver {
	return &gqlQueryResolver{resolver: resolver}
}

func (r *gqlQueryResolver) ToGitTreeLSIFData() (gql.GitTreeLSIFDataResolver, bool) { return r, true }
func (r *gqlQueryResolver) ToGitBlobLSIFData() (gql.GitBlobLSIFDataResolver, bool) { return r, true }

func (r *gqlQueryResolver) Definitions(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	locations, err := r.resolver.Definitions(ctx, int(args.Line), int(args.Character))
	if err != nil {
		return nil, err
	}

	return NewGraphQLLocationConnectionResolver(locations, nil), nil
}

func (r *gqlQueryResolver) References(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	limit := int32Default(args.First, DefaultReferencesPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}
	cursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	locations, cursor, err := r.resolver.References(ctx, int(args.Line), int(args.Character), limit, cursor)
	if err != nil {
		return nil, err
	}

	return NewGraphQLLocationConnectionResolver(locations, strPtr(cursor)), nil
}

func (r *gqlQueryResolver) Hover(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.HoverResolver, error) {
	text, lspRange, exists, err := r.resolver.Hover(ctx, int(args.Line), int(args.Character))
	if err != nil || !exists {
		return nil, err
	}

	return NewGraphQLHoverResolver(text, lspRange), nil
}

func (r *gqlQueryResolver) Diagnostics(ctx context.Context, args *gql.LSIFDiagnosticsArgs) (gql.DiagnosticConnectionResolver, error) {
	limit := int32Default(args.First, DefaultDiagnosticsPageSize)
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	diagnostics, totalCount, err := r.resolver.Diagnostics(ctx, limit)
	if err != nil {
		return nil, err
	}

	return NewGraphQLDiagnosticConnectionResolver(diagnostics, totalCount), nil
}
