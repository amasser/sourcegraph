package resolvers

import (
	"context"

	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type AdjustedDiagnostic struct {
	bundles.Diagnostic
	dump           store.Dump
	adjustedCommit string
	adjustedRange  lsp.Range
}

type diagnosticResolver struct {
	diagnostic         AdjustedDiagnostic
	collectionResolver *repositoryCollectionResolver
}

var _ gql.DiagnosticResolver = &diagnosticResolver{}

func (r *diagnosticResolver) Severity() (*string, error) { return toSeverity(r.diagnostic.Severity) }
func (r *diagnosticResolver) Code() (*string, error)     { return strPtr(r.diagnostic.Code), nil }
func (r *diagnosticResolver) Source() (*string, error)   { return strPtr(r.diagnostic.Source), nil }
func (r *diagnosticResolver) Message() (*string, error)  { return strPtr(r.diagnostic.Message), nil }

func (r *diagnosticResolver) Location(ctx context.Context) (gql.LocationResolver, error) {
	return resolveLocation(ctx, r.collectionResolver, api.RepoID(r.diagnostic.dump.RepositoryID), r.diagnostic.adjustedCommit, r.diagnostic.Path, r.diagnostic.adjustedRange)
}

//
//

func resolveLocation(ctx context.Context, collectionResolver *repositoryCollectionResolver, repositoryID api.RepoID, commit, path string, lspRange lsp.Range) (gql.LocationResolver, error) {
	treeResolver, err := collectionResolver.resolve(ctx, repositoryID, commit, path)
	if err != nil {
		return nil, err
	}

	if treeResolver == nil {
		return nil, nil
	}

	return gql.NewLocationResolver(treeResolver, &lspRange), nil
}
