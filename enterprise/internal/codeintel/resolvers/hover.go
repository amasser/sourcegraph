package resolvers

import (
	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type hoverResolver struct {
	text     string
	lspRange lsp.Range
}

var _ gql.HoverResolver = &hoverResolver{}

func (r *hoverResolver) Markdown() gql.MarkdownResolver { return gql.NewMarkdownResolver(r.text) }
func (r *hoverResolver) Range() gql.RangeResolver       { return gql.NewRangeResolver(r.lspRange) }
