package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type UploadResolver struct {
	upload store.Upload
}

func NewUploadResolver(upload store.Upload) gql.LSIFUploadResolver {
	return &UploadResolver{upload: upload}
}

func (r *UploadResolver) ID() graphql.ID            { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *UploadResolver) InputCommit() string       { return r.upload.Commit }
func (r *UploadResolver) InputRoot() string         { return r.upload.Root }
func (r *UploadResolver) IsLatestForRepo() bool     { return r.upload.VisibleAtTip }
func (r *UploadResolver) UploadedAt() gql.DateTime  { return gql.DateTime{Time: r.upload.UploadedAt} }
func (r *UploadResolver) State() string             { return strings.ToUpper(r.upload.State) }
func (r *UploadResolver) Failure() *string          { return r.upload.FailureMessage }
func (r *UploadResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.upload.StartedAt) }
func (r *UploadResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.upload.FinishedAt) }
func (r *UploadResolver) InputIndexer() string      { return r.upload.Indexer }
func (r *UploadResolver) PlaceInQueue() *int32      { return int32Ptr(r.upload.Rank) }

func (r *UploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}

//
//

type IndexResolver struct {
	index store.Index
}

func NewIndexResolver(index store.Index) gql.LSIFIndexResolver {
	return &IndexResolver{index: index}
}

func (r *IndexResolver) ID() graphql.ID            { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *IndexResolver) InputCommit() string       { return r.index.Commit }
func (r *IndexResolver) QueuedAt() gql.DateTime    { return gql.DateTime{Time: r.index.QueuedAt} }
func (r *IndexResolver) State() string             { return strings.ToUpper(r.index.State) }
func (r *IndexResolver) Failure() *string          { return r.index.FailureMessage }
func (r *IndexResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.index.StartedAt) }
func (r *IndexResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.index.FinishedAt) }
func (r *IndexResolver) PlaceInQueue() *int32      { return int32Ptr(r.index.Rank) }

func (r *IndexResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, "")
}

//
//

type HoverResolver struct {
	text     string
	lspRange lsp.Range
}

func NewHoverResolver(text string, lspRange lsp.Range) gql.HoverResolver {
	return &HoverResolver{text: text, lspRange: lspRange}
}

func (r *HoverResolver) Markdown() gql.MarkdownResolver { return gql.NewMarkdownResolver(r.text) }
func (r *HoverResolver) Range() gql.RangeResolver       { return gql.NewRangeResolver(r.lspRange) }

//
//

type DiagnosticResolver struct {
	diagnostic         resolvers.AdjustedDiagnostic
	collectionResolver *repositoryCollectionResolver
}

func NewDiagnosticResolver(diagnostic resolvers.AdjustedDiagnostic, collectionResolver *repositoryCollectionResolver) gql.DiagnosticResolver {
	return &DiagnosticResolver{diagnostic: diagnostic, collectionResolver: collectionResolver}
}

func (r *DiagnosticResolver) Severity() (*string, error) { return toSeverity(r.diagnostic.Severity) }
func (r *DiagnosticResolver) Code() (*string, error)     { return strPtr(r.diagnostic.Code), nil }
func (r *DiagnosticResolver) Source() (*string, error)   { return strPtr(r.diagnostic.Source), nil }
func (r *DiagnosticResolver) Message() (*string, error)  { return strPtr(r.diagnostic.Message), nil }

func (r *DiagnosticResolver) Location(ctx context.Context) (gql.LocationResolver, error) {
	return resolveLocation(
		ctx,
		r.collectionResolver,
		resolvers.AdjustedLocation{
			Dump:           r.diagnostic.Dump,
			Path:           r.diagnostic.AdjustedCommit,
			AdjustedCommit: r.diagnostic.Path,
			AdjustedRange:  r.diagnostic.AdjustedRange,
		},
	)
}
