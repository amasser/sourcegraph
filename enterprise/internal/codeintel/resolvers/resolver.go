package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Resolver struct {
	resolver *realResolver
}

var _ gql.CodeIntelResolver = &Resolver{}

func NewResolver(store store.Store, bundleManagerClient bundles.BundleManagerClient, codeIntelAPI codeintelapi.CodeIntelAPI) gql.CodeIntelResolver {
	return &Resolver{resolver: &realResolver{
		store:               store,
		bundleManagerClient: bundleManagerClient,
		codeIntelAPI:        codeIntelAPI,
	}}
}

func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (gql.LSIFUploadResolver, error) {
	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	upload, exists, err := r.resolver.GetUploadByID(ctx, int(uploadID))
	if err != nil || !exists {
		return nil, err
	}

	return &uploadResolver{upload}, nil
}

func (r *Resolver) LSIFUploads(ctx context.Context, args *gql.LSIFUploadsQueryArgs) (gql.LSIFUploadConnectionResolver, error) {
	return r.LSIFUploadsByRepo(ctx, &gql.LSIFRepositoryUploadsQueryArgs{LSIFUploadsQueryArgs: args})
}

func (r *Resolver) LSIFUploadsByRepo(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (gql.LSIFUploadConnectionResolver, error) {
	opts, err := toGetUploadsOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	return &uploadConnectionResolver{resolver: r.resolver.UploadConnectionResolver(opts)}, nil
}

func (r *Resolver) DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*gql.EmptyResponse, error) {
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

func (r *Resolver) LSIFIndexByID(ctx context.Context, id graphql.ID) (gql.LSIFIndexResolver, error) {
	indexID, err := unmarshalLSIFIndexGQLID(id)
	if err != nil {
		return nil, err
	}

	index, exists, err := r.resolver.GetIndexByID(ctx, int(indexID))
	if err != nil || !exists {
		return nil, err
	}

	return &indexResolver{index}, nil
}

func (r *Resolver) LSIFIndexes(ctx context.Context, args *gql.LSIFIndexesQueryArgs) (gql.LSIFIndexConnectionResolver, error) {
	return r.LSIFIndexesByRepo(ctx, &gql.LSIFRepositoryIndexesQueryArgs{LSIFIndexesQueryArgs: args})
}

func (r *Resolver) LSIFIndexesByRepo(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (gql.LSIFIndexConnectionResolver, error) {
	opts, err := toGetIndexesOptions(ctx, args)
	if err != nil {
		return nil, err
	}

	return &indexConnectionResolver{resolver: r.resolver.IndexConnectionResolver(opts)}, nil
}

func (r *Resolver) DeleteLSIFIndex(ctx context.Context, id graphql.ID) (*gql.EmptyResponse, error) {
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

func (r *Resolver) GitBlobLSIFData(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (gql.GitBlobLSIFDataResolver, error) {
	resolver, err := r.resolver.QueryResolver(ctx, args)
	if err != nil || resolver == nil {
		return nil, err
	}

	return &lsifQueryResolver{resolver: resolver}, nil
}

//
//

func toGetUploadsOptions(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (store.GetUploadsOptions, error) {
	repositoryID, err := resolveRepositoryID(ctx, args.RepositoryID)
	if err != nil {
		return store.GetUploadsOptions{}, err
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return store.GetUploadsOptions{}, err
	}

	return store.GetUploadsOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(strDefault(args.State, "")),
		Term:         strDefault(args.Query, ""),
		VisibleAtTip: boolDefault(args.IsLatestForRepo, false),
		Limit:        int32Default(args.First, DefaultUploadPageSize),
		Offset:       offset,
	}, nil
}

func toGetIndexesOptions(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (store.GetIndexesOptions, error) {
	repositoryID, err := resolveRepositoryID(ctx, args.RepositoryID)
	if err != nil {
		return store.GetIndexesOptions{}, err
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return store.GetIndexesOptions{}, err
	}

	return store.GetIndexesOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(strDefault(args.State, "")),
		Term:         strDefault(args.Query, ""),
		Limit:        int32Default(args.First, DefaultIndexPageSize),
		Offset:       offset,
	}, nil
}

func resolveRepositoryID(ctx context.Context, id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repositoryResolver, err := gql.RepositoryByID(ctx, id)
	if err != nil {
		return 0, err
	}

	return int(repositoryResolver.Type().ID), nil
}

func decodeIntCursor(val *string) (int, error) {
	if val == nil {
		return 0, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return 0, err
	}

	v, _ := strconv.Atoi(string(decoded))
	return v, nil
}

func encodeIntCursor(val *int) *graphqlutil.PageInfo {
	if val != nil {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", val))))
	}

	return graphqlutil.HasNextPage(false)
}

func strDefault(val *string, defaultValue string) string {
	if val != nil {
		return *val
	}
	return defaultValue
}

func int32Default(val *int32, defaultValue int) int {
	if val != nil {
		return int(*val)
	}
	return defaultValue
}

func boolDefault(val *bool, defaultValue bool) bool {
	return (val != nil && *val) || defaultValue
}

//
//

type realResolver struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI
}

func (r *realResolver) GetUploadByID(ctx context.Context, id int) (store.Upload, bool, error) {
	return r.store.GetUploadByID(ctx, id)
}

func (r *realResolver) GetIndexByID(ctx context.Context, id int) (store.Index, bool, error) {
	return r.store.GetIndexByID(ctx, id)
}

func (r *realResolver) UploadConnectionResolver(opts store.GetUploadsOptions) *realLsifUploadConnectionResolver {
	return &realLsifUploadConnectionResolver{store: r.store, opts: opts}
}

func (r *realResolver) IndexConnectionResolver(opts store.GetIndexesOptions) *realLsifIndexConnectionResolver {
	return &realLsifIndexConnectionResolver{store: r.store, opts: opts}
}

func (r *realResolver) DeleteUploadByID(ctx context.Context, uploadID int) error {
	getTipCommit := func(repositoryID int) (string, error) {
		tipCommit, err := gitserver.Head(ctx, r.store, repositoryID)
		if err != nil {
			return "", errors.Wrap(err, "gitserver.Head")
		}
		return tipCommit, nil
	}

	_, err := r.store.DeleteUploadByID(ctx, uploadID, getTipCommit) // TODO - modify this type to take a context
	return err
}

func (r *realResolver) DeleteIndexByID(ctx context.Context, id int) error {
	_, err := r.store.DeleteIndexByID(ctx, id)
	return err
}

func (r *realResolver) QueryResolver(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (*realLsifQueryResolver, error) {
	dumps, err := r.codeIntelAPI.FindClosestDumps(ctx, int(args.Repository.Type().ID), string(args.Commit), args.Path, args.ExactPath, args.ToolName)
	if err != nil || len(dumps) == 0 {
		return nil, err
	}

	return &realLsifQueryResolver{
		store:               r.store,
		bundleManagerClient: r.bundleManagerClient,
		codeIntelAPI:        r.codeIntelAPI,
		repo:                args.Repository.Type(),
		commit:              args.Commit,
		path:                args.Path,
		uploads:             dumps,
	}, nil
}
