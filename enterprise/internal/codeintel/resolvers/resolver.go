package resolvers

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
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
	// TODO - remove this type
	opts := LSIFUploadsListOptions{
		RepositoryID:    args.RepositoryID,
		Query:           args.Query,
		State:           args.State,
		IsLatestForRepo: args.IsLatestForRepo,
	}
	if args.First != nil {
		opts.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opts.NextURL = &nextURL
	}

	opts2, err := toGetUploadsOptions(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &uploadConnectionResolver{resolver: r.resolver.UploadConnectionResolver(opts2)}, nil
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
	// TODO - remove this type
	opts := LSIFIndexesListOptions{
		RepositoryID: args.RepositoryID,
		Query:        args.Query,
		State:        args.State,
	}
	if args.First != nil {
		opts.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opts.NextURL = &nextURL
	}

	opts2, err := toGetIndexesOptions(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &indexConnectionResolver{resolver: r.resolver.IndexConnectionResolver(opts2)}, nil
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

func toGetUploadsOptions(ctx context.Context, opts LSIFUploadsListOptions) (store.GetUploadsOptions, error) {
	var id int
	if opts.RepositoryID != "" {
		repositoryResolver, err := gql.RepositoryByID(ctx, opts.RepositoryID)
		if err != nil {
			return store.GetUploadsOptions{}, err
		}

		id = int(repositoryResolver.Type().ID)
	}
	query := ""

	if opts.Query != nil {
		query = *opts.Query
	}

	state := ""
	if opts.State != nil {
		state = strings.ToLower(*opts.State)
	}

	limit := DefaultUploadPageSize
	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	offset := 0
	if opts.NextURL != nil {
		offset, _ = strconv.Atoi(*opts.NextURL)
	}

	return store.GetUploadsOptions{
		RepositoryID: id,
		State:        state,
		Term:         query,
		VisibleAtTip: opts.IsLatestForRepo != nil && *opts.IsLatestForRepo,
		Limit:        limit,
		Offset:       offset,
	}, nil
}

func toGetIndexesOptions(ctx context.Context, opts LSIFIndexesListOptions) (store.GetIndexesOptions, error) {
	var id int
	if opts.RepositoryID != "" {
		repositoryResolver, err := gql.RepositoryByID(ctx, opts.RepositoryID)
		if err != nil {
			return store.GetIndexesOptions{}, err
		}

		id = int(repositoryResolver.Type().ID)
	}

	query := ""
	if opts.Query != nil {
		query = *opts.Query
	}

	state := ""
	if opts.State != nil {
		state = strings.ToLower(*opts.State)
	}

	limit := DefaultIndexPageSize
	if opts.Limit != nil {
		limit = int(*opts.Limit)
	}

	offset := 0
	if opts.NextURL != nil {
		offset, _ = strconv.Atoi(*opts.NextURL)
	}

	return store.GetIndexesOptions{
		RepositoryID: id,
		State:        state,
		Term:         query,
		Limit:        limit,
		Offset:       offset,
	}, nil
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

	_, err := r.store.DeleteUploadByID(ctx, uploadID, getTipCommit)
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
