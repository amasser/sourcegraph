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

type resolver struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI
}

func (r *resolver) GetUploadByID(ctx context.Context, id int) (store.Upload, bool, error) {
	return r.store.GetUploadByID(ctx, id)
}

func (r *resolver) GetIndexByID(ctx context.Context, id int) (store.Index, bool, error) {
	return r.store.GetIndexByID(ctx, id)
}

func (r *resolver) UploadConnectionResolver(opts store.GetUploadsOptions) *uploadsResolver {
	return NewUploadsResolver(r.store, opts)
}

func (r *resolver) IndexConnectionResolver(opts store.GetIndexesOptions) *indexesResolver {
	return NewIndexesResolver(r.store, opts)
}

func (r *resolver) DeleteUploadByID(ctx context.Context, uploadID int) error {
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

func (r *resolver) DeleteIndexByID(ctx context.Context, id int) error {
	_, err := r.store.DeleteIndexByID(ctx, id)
	return err
}

func (r *resolver) QueryResolver(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (*queryResolver, error) {
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

type uploadsResolver struct {
	store store.Store
	opts  store.GetUploadsOptions
	once  sync.Once
	//
	uploads    []store.Upload
	totalCount int
	nextOffset *int
	err        error
}

func NewUploadsResolver(store store.Store, opts store.GetUploadsOptions) *uploadsResolver {
	return &uploadsResolver{
		store: store,
		opts:  opts,
	}
}

func (r *uploadsResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *uploadsResolver) resolve(ctx context.Context) error {
	uploads, totalCount, err := r.store.GetUploads(ctx, r.opts)
	if err != nil {
		return err
	}

	r.uploads = uploads
	r.nextOffset = nextOffset(r.opts.Offset, len(uploads), totalCount)
	r.totalCount = totalCount
	return nil
}

type indexesResolver struct {
	store store.Store
	opts  store.GetIndexesOptions
	once  sync.Once
	//
	indexes    []store.Index
	totalCount int
	nextOffset *int
	err        error
}

func NewIndexesResolver(store store.Store, opts store.GetIndexesOptions) *indexesResolver {
	return &indexesResolver{
		store: store,
		opts:  opts,
	}
}

func (r *indexesResolver) Resolve(ctx context.Context) error {
	r.once.Do(func() { r.err = r.resolve(ctx) })
	return r.err
}

func (r *indexesResolver) resolve(ctx context.Context) error {
	indexes, totalCount, err := r.store.GetIndexes(ctx, r.opts)
	if err != nil {
		return err
	}

	r.indexes = indexes
	r.nextOffset = nextOffset(r.opts.Offset, len(indexes), totalCount)
	r.totalCount = totalCount
	return nil
}

func nextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}
