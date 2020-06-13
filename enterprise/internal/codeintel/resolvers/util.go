package resolvers

import (
	"github.com/sourcegraph/go-lsp"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
)

// TODO - document, move
func convertRange(r bundles.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

// TODO - document, move
func convertPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

// TODO - document
func nextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}
