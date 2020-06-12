package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type lsifUploadResolver struct {
	lsifUpload store.Upload
}

var _ gql.LSIFUploadResolver = &lsifUploadResolver{}

func (r *lsifUploadResolver) ID() graphql.ID        { return marshalLSIFUploadGQLID(int64(r.lsifUpload.ID)) }
func (r *lsifUploadResolver) InputCommit() string   { return r.lsifUpload.Commit }
func (r *lsifUploadResolver) InputRoot() string     { return r.lsifUpload.Root }
func (r *lsifUploadResolver) InputIndexer() string  { return r.lsifUpload.Indexer }
func (r *lsifUploadResolver) State() string         { return strings.ToUpper(r.lsifUpload.State) }
func (r *lsifUploadResolver) IsLatestForRepo() bool { return r.lsifUpload.VisibleAtTip }

func (r *lsifUploadResolver) PlaceInQueue() *int32 {
	if r.lsifUpload.Rank == nil {
		return nil
	}

	v := int32(*r.lsifUpload.Rank)
	return &v
}

func (r *lsifUploadResolver) UploadedAt() gql.DateTime {
	return gql.DateTime{Time: r.lsifUpload.UploadedAt}
}

func (r *lsifUploadResolver) StartedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.lsifUpload.StartedAt)
}

func (r *lsifUploadResolver) FinishedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.lsifUpload.FinishedAt)
}

func (r *lsifUploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.lsifUpload.RepositoryID), r.lsifUpload.Commit, r.lsifUpload.Root)
}

func (r *lsifUploadResolver) Failure() *string {
	return r.lsifUpload.FailureMessage
}
