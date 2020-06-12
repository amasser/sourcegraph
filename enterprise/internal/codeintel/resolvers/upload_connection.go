package resolvers

import (
	"context"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type uploadConnectionResolver struct {
	resolver *realLsifUploadConnectionResolver
}

var _ gql.LSIFUploadConnectionResolver = &uploadConnectionResolver{}

func (r *uploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	return resolveUploads(r.resolver.uploads), nil
}

func (r *uploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	return int32Ptr(&r.resolver.totalCount), nil
}

func (r *uploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	return encodeIntCursor(r.resolver.nextOffset), nil
}

//
//

func resolveUploads(uploads []store.Upload) []gql.LSIFUploadResolver {
	var resolvers []gql.LSIFUploadResolver
	for _, upload := range uploads {
		resolvers = append(resolvers, &uploadResolver{upload})
	}

	return resolvers
}

//
//

type realLsifUploadConnectionResolver struct {
	store store.Store
	opts  store.GetUploadsOptions
	once  sync.Once
	//
	uploads    []store.Upload
	totalCount int
	nextOffset *int
	err        error
}

func (r *realLsifUploadConnectionResolver) Compute(ctx context.Context) error {
	r.once.Do(func() { r.err = r.compute(ctx) })
	return r.err
}

func (r *realLsifUploadConnectionResolver) compute(ctx context.Context) error {
	uploads, totalCount, err := r.store.GetUploads(ctx, r.opts)
	if err != nil {
		return err
	}

	var nextOffset *int
	if r.opts.Offset+len(uploads) < totalCount {
		val := r.opts.Offset + len(uploads)
		nextOffset = &val
	}

	r.uploads = uploads
	r.nextOffset = nextOffset
	r.totalCount = totalCount
	return nil
}
