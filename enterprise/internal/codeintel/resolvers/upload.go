package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type uploadResolver struct {
	upload store.Upload
}

var _ gql.LSIFUploadResolver = &uploadResolver{}

func (r *uploadResolver) ID() graphql.ID        { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *uploadResolver) InputCommit() string   { return r.upload.Commit }
func (r *uploadResolver) InputRoot() string     { return r.upload.Root }
func (r *uploadResolver) InputIndexer() string  { return r.upload.Indexer }
func (r *uploadResolver) State() string         { return strings.ToUpper(r.upload.State) }
func (r *uploadResolver) IsLatestForRepo() bool { return r.upload.VisibleAtTip }

func (r *uploadResolver) PlaceInQueue() *int32 {
	if r.upload.Rank == nil {
		return nil
	}

	v := int32(*r.upload.Rank)
	return &v
}

func (r *uploadResolver) UploadedAt() gql.DateTime {
	return gql.DateTime{Time: r.upload.UploadedAt}
}

func (r *uploadResolver) StartedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.upload.StartedAt)
}

func (r *uploadResolver) FinishedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.upload.FinishedAt)
}

func (r *uploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}

func (r *uploadResolver) Failure() *string {
	return r.upload.FailureMessage
}
