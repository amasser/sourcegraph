package resolvers

import (
	"testing"
	// "github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	// apimocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api/mocks"
	// bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client/mocks"
	// "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	// storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
)

func TestDefinitions(t *testing.T) {
	// TODO - test

	// mockStore := storemocks.NewMockStore()
	// mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	// mockCodeIntelAPI := apimocks.NewMockCodeIntelAPI()

	// resolver := NewQueryResolver(
	// 	mockStore,
	// 	mockBundleManagerClient,
	// 	mockCodeIntelAPI,
	// 	NewPositionAdjuster(&types.Repo{ID: 50}, "deadbeef"),
	// 	50,
	// 	"deadbeef",
	// 	"foo/bar/baz.go",
	// 	[]store.Dump{
	// 		{ID: 1},
	// 		{ID: 2},
	// 		{ID: 3},
	// 	},
	// )

	// definitions, err := resolver.Definitions(context.Background(), 10, 15)
	// if err != nil {
	// 	t.Fatalf("unexpected error resolving definitions: %s", err)
	// }

	// fmt.Printf("> %v\n", definitions)
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
