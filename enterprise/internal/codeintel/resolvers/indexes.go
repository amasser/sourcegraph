package resolvers

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// TODO - document
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

// TODO - document
func NewIndexesResolver(store store.Store, opts store.GetIndexesOptions) *IndexesResolver {
	return &IndexesResolver{store: store, opts: opts}
}

// TODO - document
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
