package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type indexResolver struct {
	index store.Index
}

var _ gql.LSIFIndexResolver = &indexResolver{}

func (r *indexResolver) ID() graphql.ID      { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *indexResolver) InputCommit() string { return r.index.Commit }
func (r *indexResolver) State() string       { return strings.ToUpper(r.index.State) }

func (r *indexResolver) PlaceInQueue() *int32 {
	if r.index.Rank == nil {
		return nil
	}

	v := int32(*r.index.Rank)
	return &v
}

func (r *indexResolver) QueuedAt() gql.DateTime {
	return gql.DateTime{Time: r.index.QueuedAt}
}

func (r *indexResolver) StartedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.index.StartedAt)
}

func (r *indexResolver) FinishedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.index.FinishedAt)
}

func (r *indexResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, "")
}

func (r *indexResolver) Failure() *string {
	return r.index.FailureMessage
}
