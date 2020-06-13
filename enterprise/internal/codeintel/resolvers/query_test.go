package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestDefinitions(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockCodeIntelAPI := apimocks.NewMockCodeIntelAPI()

	resolver := NewQueryResolver(
		mockStore,
		mockBundleManagerClient,
		mockCodeIntelAPI,
		&types.Repo{ID: 50},
		api.CommitID("deadbeef"),
		"foo/bar/baz.go",
		[]store.Dump{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		},
	)

	definitions, err := resolver.Definitions(context.Background(), 10, 15)
	if err != nil {
		t.Fatalf("unexpected error resolving definitions: %s", err)
	}

	fmt.Printf("> %v\n", definitions)
}

func TestReferences(t *testing.T) {
	// TODO - test
}

func TestHover(t *testing.T) {
	// TODO - test
}

func TestDiagnostics(t *testing.T) {
	// TODO - test
}
