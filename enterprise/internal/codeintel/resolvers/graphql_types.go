package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type gqlUploadResolver struct {
	upload store.Upload
}

func NewGraphQLUploadResolver(upload store.Upload) gql.LSIFUploadResolver {
	return &gqlUploadResolver{
		upload: upload,
	}
}

func (r *gqlUploadResolver) ID() graphql.ID            { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *gqlUploadResolver) InputCommit() string       { return r.upload.Commit }
func (r *gqlUploadResolver) InputRoot() string         { return r.upload.Root }
func (r *gqlUploadResolver) IsLatestForRepo() bool     { return r.upload.VisibleAtTip }
func (r *gqlUploadResolver) UploadedAt() gql.DateTime  { return gql.DateTime{Time: r.upload.UploadedAt} }
func (r *gqlUploadResolver) State() string             { return strings.ToUpper(r.upload.State) }
func (r *gqlUploadResolver) Failure() *string          { return r.upload.FailureMessage }
func (r *gqlUploadResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.upload.StartedAt) }
func (r *gqlUploadResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.upload.FinishedAt) }
func (r *gqlUploadResolver) InputIndexer() string      { return r.upload.Indexer }
func (r *gqlUploadResolver) PlaceInQueue() *int32      { return int32Ptr(r.upload.Rank) }

func (r *gqlUploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}

//
//

type gqlIndexResolver struct {
	index store.Index
}

func NewGraphQLIndexResolver(index store.Index) gql.LSIFIndexResolver {
	return &gqlIndexResolver{
		index: index,
	}
}

func (r *gqlIndexResolver) ID() graphql.ID            { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *gqlIndexResolver) InputCommit() string       { return r.index.Commit }
func (r *gqlIndexResolver) QueuedAt() gql.DateTime    { return gql.DateTime{Time: r.index.QueuedAt} }
func (r *gqlIndexResolver) State() string             { return strings.ToUpper(r.index.State) }
func (r *gqlIndexResolver) Failure() *string          { return r.index.FailureMessage }
func (r *gqlIndexResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.index.StartedAt) }
func (r *gqlIndexResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.index.FinishedAt) }
func (r *gqlIndexResolver) PlaceInQueue() *int32      { return int32Ptr(r.index.Rank) }

func (r *gqlIndexResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, "")
}

//
//

type gqlHoverResolver struct {
	text     string
	lspRange lsp.Range
}

func NewGraphQLHoverResolver(text string, lspRange lsp.Range) gql.HoverResolver {
	return &gqlHoverResolver{
		text:     text,
		lspRange: lspRange,
	}
}

func (r *gqlHoverResolver) Markdown() gql.MarkdownResolver { return gql.NewMarkdownResolver(r.text) }
func (r *gqlHoverResolver) Range() gql.RangeResolver       { return gql.NewRangeResolver(r.lspRange) }

//
//

type gqlDiagnosticResolver struct {
	diagnostic         AdjustedDiagnostic
	collectionResolver *repositoryCollectionResolver
}

func NewGraphQLDiagnosticResolver(diagnostic AdjustedDiagnostic, collectionResolver *repositoryCollectionResolver) gql.DiagnosticResolver {
	return &gqlDiagnosticResolver{
		diagnostic:         diagnostic,
		collectionResolver: collectionResolver,
	}
}

func (r *gqlDiagnosticResolver) Severity() (*string, error) { return toSeverity(r.diagnostic.Severity) }
func (r *gqlDiagnosticResolver) Code() (*string, error)     { return strPtr(r.diagnostic.Code), nil }
func (r *gqlDiagnosticResolver) Source() (*string, error)   { return strPtr(r.diagnostic.Source), nil }
func (r *gqlDiagnosticResolver) Message() (*string, error)  { return strPtr(r.diagnostic.Message), nil }

func (r *gqlDiagnosticResolver) Location(ctx context.Context) (gql.LocationResolver, error) {
	return resolveLocation(
		ctx,
		r.collectionResolver,
		AdjustedLocation{
			dump:           r.diagnostic.dump,
			path:           r.diagnostic.adjustedCommit,
			adjustedCommit: r.diagnostic.Path,
			adjustedRange:  r.diagnostic.adjustedRange,
		},
	)
}
