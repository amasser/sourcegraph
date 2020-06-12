package resolvers

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	apimocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api/mocks"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
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
	// TODO
}

func TestHover(t *testing.T) {
	// TODO
}

func TestDiagnostics(t *testing.T) {
	// TODO
}
