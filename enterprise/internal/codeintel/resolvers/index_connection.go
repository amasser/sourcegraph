package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
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
	if r.resolver.totalCount == nil {
		return nil, nil
	}

	c := int32(*r.resolver.totalCount)
	return &c, nil
}

func (r *indexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.compute(ctx); err != nil {
		return nil, err
	}

	if r.resolver.nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(r.resolver.nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

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
	totalCount *int
	nextURL    string
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

	cursor := ""
	if r.opts.Offset+len(indexes) < totalCount {
		cursor = fmt.Sprintf("%d", r.opts.Offset+len(indexes))
	}

	r.indexes = indexes
	r.nextURL = cursor
	r.totalCount = &totalCount
	return nil
}
