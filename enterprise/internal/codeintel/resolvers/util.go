package resolvers

import (
	"github.com/sourcegraph/go-lsp"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
)

// convertRange creates an LSP range from a bundle range.
func convertRange(r bundles.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

// convertPosition creates an LSP position from a line and character pair.
func convertPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

// nextOffset determines the offset that should be used for a subsequent request.
// If there are no more results in the paged result set, this function returns nil.
func nextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}
