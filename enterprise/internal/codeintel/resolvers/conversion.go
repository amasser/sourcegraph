package resolvers

import (
	"fmt"

	"github.com/sourcegraph/go-lsp"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
)

func convertRange(r bundles.Range) lsp.Range {
	return lsp.Range{
		Start: lsp.Position{Line: r.Start.Line, Character: r.Start.Character},
		End:   lsp.Position{Line: r.End.Line, Character: r.End.Character},
	}
}

var severities = map[int]string{
	1: "ERROR",
	2: "WARNING",
	3: "INFORMATION",
	4: "HINT",
}

func toSeverity(val int) (*string, error) {
	severity, ok := severities[val]
	if !ok {
		return nil, fmt.Errorf("unknown diagnostic severity %d", val)
	}

	return &severity, nil
}

func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

func int32Ptr(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}
