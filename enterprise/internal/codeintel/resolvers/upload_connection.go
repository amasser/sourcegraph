package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
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
	if r.resolver.totalCount == nil {
		return nil, nil
	}

	c := int32(*r.resolver.totalCount)
	return &c, nil
}

func (r *uploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	if r.resolver.nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(r.resolver.nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func resolveUploads(uploads []store.Upload) []gql.LSIFUploadResolver {
	var resolvers []gql.LSIFUploadResolver
	for _, upload := range uploads {
		resolvers = append(resolvers, &uploadResolver{upload})
	}

	return resolvers
}

//
//

type LSIFUploadsListOptions struct {
	RepositoryID    graphql.ID
	Query           *string
	State           *string
	IsLatestForRepo *bool
	Limit           *int32
	NextURL         *string
}

type realLsifUploadConnectionResolver struct {
	store store.Store
	opts  store.GetUploadsOptions
	once  sync.Once
	//
	uploads    []store.Upload
	totalCount *int
	nextURL    string
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

	cursor := ""
	if r.opts.Offset+len(uploads) < totalCount {
		cursor = fmt.Sprintf("%d", r.opts.Offset+len(uploads))
	}

	r.uploads = uploads
	r.nextURL = cursor
	r.totalCount = &totalCount
	return nil
}
