package resolvers

import (
	"context"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type indexConnectionResolver struct {
	resolver *realLsifIndexConnectionResolver
}

var _ gql.LSIFIndexConnectionResolver = &indexConnectionResolver{}

func (r *indexConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFIndexResolver, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	return resolveIndexes(r.resolver.indexes), nil
}

func (r *indexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	return int32Ptr(&r.resolver.totalCount), nil
}

func (r *indexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.compute(ctx); err != nil {
		return nil, err
	}

	return encodeIntCursor(r.resolver.nextOffset), nil
}

//
//

func resolveIndexes(indexes []store.Index) []gql.LSIFIndexResolver {
	var resolvers []gql.LSIFIndexResolver
	for _, index := range indexes {
		resolvers = append(resolvers, &indexResolver{index})
	}

	return resolvers
}

//
//

type realLsifIndexConnectionResolver struct {
	store store.Store
	opts  store.GetIndexesOptions
	once  sync.Once
	//
	indexes    []store.Index
	totalCount int
	nextOffset *int
	err        error
}

func (r *realLsifIndexConnectionResolver) Compute(ctx context.Context) error {
	r.once.Do(func() { r.err = r.compute(ctx) })
	return r.err
}

func (r *realLsifIndexConnectionResolver) compute(ctx context.Context) error {
	indexes, totalCount, err := r.store.GetIndexes(ctx, r.opts)
	if err != nil {
		return err
	}

	var nextOffset *int
	if r.opts.Offset+len(indexes) < totalCount {
		v := r.opts.Offset + len(indexes)
		nextOffset = &v
	}

	r.indexes = indexes
	r.nextOffset = nextOffset
	r.totalCount = totalCount
	return nil
}
