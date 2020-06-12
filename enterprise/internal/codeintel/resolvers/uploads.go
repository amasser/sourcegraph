package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

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
