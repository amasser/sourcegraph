package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

const DefaultUploadPageSize = 50
const DefaultIndexPageSize = 50

func toGetUploadsOptions(ctx context.Context, args *gql.LSIFRepositoryUploadsQueryArgs) (store.GetUploadsOptions, error) {
	repositoryID, err := resolveRepositoryID(ctx, args.RepositoryID)
	if err != nil {
		return store.GetUploadsOptions{}, err
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return store.GetUploadsOptions{}, err
	}

	return store.GetUploadsOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(strDefault(args.State, "")),
		Term:         strDefault(args.Query, ""),
		VisibleAtTip: boolDefault(args.IsLatestForRepo, false),
		Limit:        int32Default(args.First, DefaultUploadPageSize),
		Offset:       offset,
	}, nil
}

func toGetIndexesOptions(ctx context.Context, args *gql.LSIFRepositoryIndexesQueryArgs) (store.GetIndexesOptions, error) {
	repositoryID, err := resolveRepositoryID(ctx, args.RepositoryID)
	if err != nil {
		return store.GetIndexesOptions{}, err
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return store.GetIndexesOptions{}, err
	}

	return store.GetIndexesOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(strDefault(args.State, "")),
		Term:         strDefault(args.Query, ""),
		Limit:        int32Default(args.First, DefaultIndexPageSize),
		Offset:       offset,
	}, nil
}

func resolveRepositoryID(ctx context.Context, id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repositoryResolver, err := gql.RepositoryByID(ctx, id)
	if err != nil {
		return 0, err
	}

	return int(repositoryResolver.Type().ID), nil
}
