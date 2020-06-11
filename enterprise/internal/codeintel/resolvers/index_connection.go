package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type lsifIndexConnectionResolver struct {
	resolver *realLsifIndexConnectionResolver
}

var _ graphqlbackend.LSIFIndexConnectionResolver = &lsifIndexConnectionResolver{}

type LSIFIndexesListOptions struct {
	RepositoryID graphql.ID
	Query        *string
	State        *string
	Limit        *int32
	NextURL      *string
}

func (r *lsifIndexConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFIndexResolver, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}

	var l []graphqlbackend.LSIFIndexResolver
	for _, lsifIndex := range r.resolver.indexes {
		l = append(l, &lsifIndexResolver{
			lsifIndex: lsifIndex,
		})
	}
	return l, nil
}

func (r *lsifIndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Compute(ctx); err != nil {
		return nil, err
	}
	if r.resolver.totalCount == nil {
		return nil, nil
	}

	c := int32(*r.resolver.totalCount)
	return &c, nil
}

func (r *lsifIndexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.compute(ctx); err != nil {
		return nil, err
	}

	if r.resolver.nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(r.resolver.nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

type realLsifIndexConnectionResolver struct {
	store store.Store
	opt   LSIFIndexesListOptions
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
	var id int
	if r.opt.RepositoryID != "" {
		repositoryResolver, err := graphqlbackend.RepositoryByID(ctx, r.opt.RepositoryID)
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

	limit := DefaultIndexPageSize
	if r.opt.Limit != nil {
		limit = int(*r.opt.Limit)
	}

	offset := 0
	if r.opt.NextURL != nil {
		offset, _ = strconv.Atoi(*r.opt.NextURL)
	}

	indexes, totalCount, err := r.store.GetIndexes(ctx, store.GetIndexesOptions{
		RepositoryID: id,
		State:        state,
		Term:         query,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return err
	}

	cursor := ""
	if offset+len(indexes) < totalCount {
		cursor = fmt.Sprintf("%d", offset+len(indexes))
	}

	is := indexes

	r.indexes = is
	r.nextURL = cursor
	r.totalCount = &totalCount
	return nil
}
