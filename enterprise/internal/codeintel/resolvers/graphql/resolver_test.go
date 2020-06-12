package graphql

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/google/go-cmp/cmp"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// TODO - test resolver

func TestMakeGetUploadsOptions(t *testing.T) {
	t.Cleanup(func() {
		db.Mocks.Repos.Get = nil
	})
	db.Mocks.Repos.Get = func(v0 context.Context, id api.RepoID) (*types.Repo, error) {
		if id != 50 {
			t.Errorf("unexpected repository name. want=%d have=%d", 50, id)
		}
		return &types.Repo{ID: 50}, nil
	}

	opts, err := makeGetUploadsOptions(context.Background(), &gql.LSIFRepositoryUploadsQueryArgs{
		LSIFUploadsQueryArgs: &gql.LSIFUploadsQueryArgs{
			ConnectionArgs: graphqlutil.ConnectionArgs{
				First: intPtr(5),
			},
			Query:           strPtr("q"),
			State:           strPtr("s"),
			IsLatestForRepo: boolPtr(true),
			After:           encodeIntCursor(intPtr(25)).EndCursor(),
		},
		RepositoryID: graphql.ID(base64.StdEncoding.EncodeToString([]byte("Repo:50"))),
	})
	if err != nil {
		t.Fatalf("unexpected error making options: %s", err)
	}

	expected := store.GetUploadsOptions{
		RepositoryID: 50,
		State:        "s",
		Term:         "q",
		VisibleAtTip: true,
		Limit:        5,
		Offset:       25,
	}
	if diff := cmp.Diff(expected, opts); diff != "" {
		t.Errorf("unexpected opts (-want +got):\n%s", diff)
	}
}

func TestMakeGetUploadsOptionsDefaults(t *testing.T) {
	opts, err := makeGetUploadsOptions(context.Background(), &gql.LSIFRepositoryUploadsQueryArgs{
		LSIFUploadsQueryArgs: &gql.LSIFUploadsQueryArgs{},
	})
	if err != nil {
		t.Fatalf("unexpected error making options: %s", err)
	}

	expected := store.GetUploadsOptions{
		RepositoryID: 0,
		State:        "",
		Term:         "",
		VisibleAtTip: false,
		Limit:        DefaultUploadPageSize,
		Offset:       0,
	}
	if diff := cmp.Diff(expected, opts); diff != "" {
		t.Errorf("unexpected opts (-want +got):\n%s", diff)
	}
}

func TestMakeGetIndexesOptions(t *testing.T) {
	t.Cleanup(func() {
		db.Mocks.Repos.Get = nil
	})
	db.Mocks.Repos.Get = func(v0 context.Context, id api.RepoID) (*types.Repo, error) {
		if id != 50 {
			t.Errorf("unexpected repository name. want=%d have=%d", 50, id)
		}
		return &types.Repo{ID: 50}, nil
	}

	opts, err := makeGetIndexesOptions(context.Background(), &gql.LSIFRepositoryIndexesQueryArgs{
		LSIFIndexesQueryArgs: &gql.LSIFIndexesQueryArgs{
			ConnectionArgs: graphqlutil.ConnectionArgs{
				First: intPtr(5),
			},
			Query: strPtr("q"),
			State: strPtr("s"),
			After: encodeIntCursor(intPtr(25)).EndCursor(),
		},
		RepositoryID: graphql.ID(base64.StdEncoding.EncodeToString([]byte("Repo:50"))),
	})
	if err != nil {
		t.Fatalf("unexpected error making options: %s", err)
	}

	expected := store.GetIndexesOptions{
		RepositoryID: 50,
		State:        "s",
		Term:         "q",
		Limit:        5,
		Offset:       25,
	}
	if diff := cmp.Diff(expected, opts); diff != "" {
		t.Errorf("unexpected opts (-want +got):\n%s", diff)
	}
}

func TestMakeGetIndexesOptionsDefaults(t *testing.T) {
	opts, err := makeGetIndexesOptions(context.Background(), &gql.LSIFRepositoryIndexesQueryArgs{
		LSIFIndexesQueryArgs: &gql.LSIFIndexesQueryArgs{},
	})
	if err != nil {
		t.Fatalf("unexpected error making options: %s", err)
	}

	expected := store.GetIndexesOptions{
		RepositoryID: 0,
		State:        "",
		Term:         "",
		Limit:        DefaultIndexPageSize,
		Offset:       0,
	}
	if diff := cmp.Diff(expected, opts); diff != "" {
		t.Errorf("unexpected opts (-want +got):\n%s", diff)
	}
}
