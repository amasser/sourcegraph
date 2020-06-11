package resolvers

import (
	"context"
	"fmt"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type AdjustedDiagnostic struct {
	diagnostic     codeintelapi.ResolvedDiagnostic
	adjustedCommit string
	adjustedRange  lsp.Range
}

type diagnosticResolver struct {
	repo               *types.Repo
	diagnostic         AdjustedDiagnostic
	collectionResolver *repositoryCollectionResolver
}

var _ graphqlbackend.DiagnosticResolver = &diagnosticResolver{}

func (r *diagnosticResolver) Location(ctx context.Context) (graphqlbackend.LocationResolver, error) {
	treeResolver, err := r.collectionResolver.resolve(ctx, api.RepoID(r.repo.ID), string(r.diagnostic.adjustedCommit), r.diagnostic.diagnostic.Diagnostic.Path)
	if err != nil {
		return nil, err
	}

	if treeResolver == nil {
		return nil, nil
	}

	return graphqlbackend.NewLocationResolver(treeResolver, &r.diagnostic.adjustedRange), nil
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
