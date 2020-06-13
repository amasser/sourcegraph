package resolvers

import (
	"context"

	"github.com/pkg/errors"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Resolver struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI
}

func NewResolver(store store.Store, bundleManagerClient bundles.BundleManagerClient, codeIntelAPI codeintelapi.CodeIntelAPI) *Resolver {
	return &Resolver{
		store:               store,
		bundleManagerClient: bundleManagerClient,
		codeIntelAPI:        codeIntelAPI,
	}
}

func (r *Resolver) GetUploadByID(ctx context.Context, id int) (store.Upload, bool, error) {
	return r.store.GetUploadByID(ctx, id)
}

func (r *Resolver) GetIndexByID(ctx context.Context, id int) (store.Index, bool, error) {
	return r.store.GetIndexByID(ctx, id)
}

func (r *Resolver) UploadConnectionResolver(opts store.GetUploadsOptions) *UploadsResolver {
	return NewUploadsResolver(r.store, opts)
}

func (r *Resolver) IndexConnectionResolver(opts store.GetIndexesOptions) *IndexesResolver {
	return NewIndexesResolver(r.store, opts)
}

func (r *Resolver) DeleteUploadByID(ctx context.Context, uploadID int) error {
	_, err := r.store.DeleteUploadByID(ctx, uploadID, r.getTipCommit)
	return err
}

func (r *Resolver) DeleteIndexByID(ctx context.Context, id int) error {
	_, err := r.store.DeleteIndexByID(ctx, id)
	return err
}

// TODO - document
func (r *Resolver) QueryResolver(ctx context.Context, args *gql.GitBlobLSIFDataArgs) (*QueryResolver, error) {
	repo := args.Repository.Type()
	repositoryID := int(repo.ID)
	commit := string(args.Commit)
	path := args.Path

	dumps, err := r.codeIntelAPI.FindClosestDumps(ctx, repositoryID, commit, path, args.ExactPath, args.ToolName)
	if err != nil || len(dumps) == 0 {
		return nil, err
	}

	positionAdjuster := NewPositionAdjuster(repo, commit)
	return NewQueryResolver(r.store, r.bundleManagerClient, r.codeIntelAPI, positionAdjuster, repositoryID, commit, path, dumps), nil
}

// TODO - document
func (r *Resolver) getTipCommit(ctx context.Context, repositoryID int) (string, error) {
	tipCommit, err := gitserver.Head(ctx, r.store, repositoryID)
	if err != nil {
		return "", errors.Wrap(err, "gitserver.Head")
	}

	return tipCommit, nil
}
