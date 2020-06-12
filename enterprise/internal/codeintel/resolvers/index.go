package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type lsifIndexResolver struct {
	lsifIndex store.Index
}

var _ gql.LSIFIndexResolver = &lsifIndexResolver{}

func (r *lsifIndexResolver) ID() graphql.ID      { return marshalLSIFIndexGQLID(int64(r.lsifIndex.ID)) }
func (r *lsifIndexResolver) InputCommit() string { return r.lsifIndex.Commit }
func (r *lsifIndexResolver) State() string       { return strings.ToUpper(r.lsifIndex.State) }

func (r *lsifIndexResolver) PlaceInQueue() *int32 {
	if r.lsifIndex.Rank == nil {
		return nil
	}

	v := int32(*r.lsifIndex.Rank)
	return &v
}

func (r *lsifIndexResolver) QueuedAt() gql.DateTime {
	return gql.DateTime{Time: r.lsifIndex.QueuedAt}
}

func (r *lsifIndexResolver) StartedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.lsifIndex.StartedAt)
}

func (r *lsifIndexResolver) FinishedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.lsifIndex.FinishedAt)
}

func (r *lsifIndexResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.lsifIndex.RepositoryID), r.lsifIndex.Commit, "")
}

func (r *lsifIndexResolver) Failure() *string {
	return r.lsifIndex.FailureMessage
}
