package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type AdjustedDiagnostic struct {
	diagnostic     codeintelapi.ResolvedDiagnostic
	adjustedCommit string
	adjustedRange  lsp.Range
}

type diagnosticResolver struct {
	diagnostic         AdjustedDiagnostic
	collectionResolver *repositoryCollectionResolver
}

var _ graphqlbackend.DiagnosticResolver = &diagnosticResolver{}

func (r *diagnosticResolver) Location(ctx context.Context) (graphqlbackend.LocationResolver, error) {
	return resolveLocation(ctx, r.collectionResolver, api.RepoID(r.diagnostic.diagnostic.Dump.RepositoryID), r.diagnostic.adjustedCommit, r.diagnostic.diagnostic.Diagnostic.Path, r.diagnostic.adjustedRange)
}

func resolveLocation(ctx context.Context, collectionResolver *repositoryCollectionResolver, repositoryID api.RepoID, commit, path string, lspRange lsp.Range) (graphqlbackend.LocationResolver, error) {
	treeResolver, err := collectionResolver.resolve(ctx, repositoryID, commit, path)
	if err != nil {
		return nil, err
	}

	if treeResolver == nil {
		return nil, nil
	}

	return graphqlbackend.NewLocationResolver(treeResolver, &lspRange), nil
}

var severities = map[int]string{
	1: "ERROR",
	2: "WARNING",
	3: "INFORMATION",
	4: "HINT",
}

func (r *diagnosticResolver) Severity(ctx context.Context) (*string, error) {
	if severity, ok := severities[r.diagnostic.diagnostic.Diagnostic.Severity]; ok {
		return &severity, nil
	}

	return nil, fmt.Errorf("unknown diagnostic severity %d", r.diagnostic.diagnostic.Diagnostic.Severity)
}

func (r *diagnosticResolver) Code(ctx context.Context) (*string, error) {
	if r.diagnostic.diagnostic.Diagnostic.Code == "" {
		return nil, nil
	}

	return &r.diagnostic.diagnostic.Diagnostic.Code, nil
}

func (r *diagnosticResolver) Source(ctx context.Context) (*string, error) {
	if r.diagnostic.diagnostic.Diagnostic.Source == "" {
		return nil, nil
	}

	return &r.diagnostic.diagnostic.Diagnostic.Source, nil
}

func (r *diagnosticResolver) Message(ctx context.Context) (*string, error) {
	if r.diagnostic.diagnostic.Diagnostic.Message == "" {
		return nil, nil
	}

	return &r.diagnostic.diagnostic.Diagnostic.Message, nil
}
