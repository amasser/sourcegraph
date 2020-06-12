package resolvers

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Resolver struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI
}

func NewResolver(store store.Store, bundleManagerClient bundles.BundleManagerClient, codeIntelAPI codeintelapi.CodeIntelAPI) *Resolver {
	return &Resolver{store: store, bundleManagerClient: bundleManagerClient, codeIntelAPI: codeIntelAPI}
}

func (r *Resolver) GetUploadByID(ctx context.Context, id int) (store.Upload, bool, error) {
	return r.store.GetUploadByID(ctx, id)
}

func (r *Resolver) GetIndexByID(ctx context.Context, id int) (store.Index, bool, error) {
	return r.store.GetIndexByID(ctx, id)
}

func (r *Resolver) UploadConnectionResolver(opts store.GetUploadsOptions) *UploadsResolver {
	return NewUploadsResolver(r.store, opts)
}

func (r *Resolver) IndexConnectionResolver(opts store.GetIndexesOptions) *IndexesResolver {
	return NewIndexesResolver(r.store, opts)
}

func (r *Resolver) DeleteUploadByID(ctx context.Context, uploadID int) error {
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

func (r *Resolver) DeleteIndexByID(ctx context.Context, id int) error {
	_, err := r.store.DeleteIndexByID(ctx, id)
	return err
}

func (r *Resolver) QueryResolver(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (*QueryResolver, error) {
	dumps, err := r.codeIntelAPI.FindClosestDumps(ctx, int(args.Repository.Type().ID), string(args.Commit), args.Path, args.ExactPath, args.ToolName)
	if err != nil || len(dumps) == 0 {
		return nil, err
	}

	return NewQueryResolver(
		r.store,
		r.bundleManagerClient,
		r.codeIntelAPI,
		args.Repository.Type(),
		args.Commit,
		args.Path,
		dumps,
	), nil
}

type UploadsResolver struct {
	store store.Store
	opts  store.GetUploadsOptions
	once  sync.Once
	//
	Uploads    []store.Upload
	TotalCount int
	NextOffset *int
	err        error
}

func NewUploadsResolver(store store.Store, opts store.GetUploadsOptions) *UploadsResolver {
	return &UploadsResolver{store: store, opts: opts}
}

func (r *UploadsResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *UploadsResolver) resolve(ctx context.Context) error {
	uploads, totalCount, err := r.store.GetUploads(ctx, r.opts)
	if err != nil {
		return err
	}

	r.Uploads = uploads
	r.NextOffset = nextOffset(r.opts.Offset, len(uploads), totalCount)
	r.TotalCount = totalCount
	return nil
}

type IndexesResolver struct {
	store store.Store
	opts  store.GetIndexesOptions
	once  sync.Once
	//
	Indexes    []store.Index
	TotalCount int
	NextOffset *int
	err        error
}

func NewIndexesResolver(store store.Store, opts store.GetIndexesOptions) *IndexesResolver {
	return &IndexesResolver{store: store, opts: opts}
}

func (r *IndexesResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *IndexesResolver) resolve(ctx context.Context) error {
	indexes, totalCount, err := r.store.GetIndexes(ctx, r.opts)
	if err != nil {
		return err
	}

	r.Indexes = indexes
	r.NextOffset = nextOffset(r.opts.Offset, len(indexes), totalCount)
	r.TotalCount = totalCount
	return nil
}

func nextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}
