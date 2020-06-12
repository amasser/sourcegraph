package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
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
	opt   LSIFUploadsListOptions
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
	var id int
	if r.opt.RepositoryID != "" {
		repositoryResolver, err := gql.RepositoryByID(ctx, r.opt.RepositoryID)
		if err != nil {
			return err
		}

		id = int(repositoryResolver.Type().ID)
	}
	query := ""

	if r.opt.Query != nil {
		query = *r.opt.Query
	}

	state := ""
	if r.opt.State != nil {
		state = strings.ToLower(*r.opt.State)
	}

	limit := DefaultUploadPageSize
	if r.opt.Limit != nil {
		limit = int(*r.opt.Limit)
	}

	offset := 0
	if r.opt.NextURL != nil {
		offset, _ = strconv.Atoi(*r.opt.NextURL)
	}

	uploads, totalCount, err := r.store.GetUploads(ctx, store.GetUploadsOptions{
		RepositoryID: id,
		State:        state,
		Term:         query,
		VisibleAtTip: r.opt.IsLatestForRepo != nil && *r.opt.IsLatestForRepo,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return err
	}

	cursor := ""
	if offset+len(uploads) < totalCount {
		cursor = fmt.Sprintf("%d", offset+len(uploads))
	}

	us := uploads

	r.uploads = us
	r.nextURL = cursor
	r.totalCount = &totalCount
	return nil
}
