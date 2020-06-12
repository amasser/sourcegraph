package resolvers

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/go-lsp"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type AdjustedLocation struct {
	Dump           store.Dump
	Path           string
	AdjustedCommit string
	AdjustedRange  lsp.Range
}

type AdjustedDiagnostic struct {
	bundles.Diagnostic
	Dump           store.Dump
	AdjustedCommit string
	AdjustedRange  lsp.Range
}

type QueryResolver struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI
	positionAdjuster    PositionAdjuster
	repositoryID        int
	commit              string
	path                string
	uploads             []store.Dump
}

func NewQueryResolver(
	store store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	codeIntelAPI codeintelapi.CodeIntelAPI,
	positionAdjuster PositionAdjuster,
	repositoryID int,
	commit string,
	uploads []store.Dump,
) *QueryResolver {
	return &QueryResolver{
		store:               store,
		bundleManagerClient: bundleManagerClient,
		codeIntelAPI:        codeIntelAPI,
		positionAdjuster:    positionAdjuster,
		repositoryID:        repositoryID,
		commit:              commit,
		uploads:             uploads,
	}
}

func (r *QueryResolver) Definitions(ctx context.Context, line, character int) ([]AdjustedLocation, error) {
	for i := range r.uploads {
		adjustedPath, adjustedLine, adjustedCharacter, ok, err := r.positionAdjuster.AdjustPosition(ctx, r.uploads[i].Commit, line, character)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		locations, err := r.codeIntelAPI.Definitions(ctx, adjustedPath, adjustedLine, adjustedCharacter, r.uploads[i].ID)
		if err != nil {
			return nil, err
		}
		if len(locations) == 0 {
			continue
		}

		adjustedLocations := make([]AdjustedLocation, 0, len(locations))
		for i := range locations {
			adjustedCommit := locations[i].Dump.Commit
			adjustedRange := convertRange(locations[i].Range)
			if locations[i].Dump.RepositoryID == r.repositoryID {
				var err error
				adjustedCommit, adjustedRange, err = r.positionAdjuster.AdjustLocation(ctx, locations[i].Dump.Commit, locations[i].Path, locations[i].Range)
				if err != nil {
					return nil, err
				}
			}

			adjustedLocations = append(adjustedLocations, AdjustedLocation{
				Dump:           locations[i].Dump,
				Path:           locations[i].Path,
				AdjustedCommit: adjustedCommit,
				AdjustedRange:  adjustedRange,
			})
		}

		return adjustedLocations, nil
	}

	return nil, nil
}

func (r *QueryResolver) References(ctx context.Context, line, character, limit int, rawCursor string) ([]AdjustedLocation, string, error) {
	// Decode a map of upload ids to the next url that serves
	// the new page of results. This may not include an entry
	// for every upload if their result sets have already been
	// exhausted.
	cursors, err := readCursor(rawCursor)
	if err != nil {
		return nil, "", err
	}

	// We need to maintain a symmetric map for the next page
	// of results that we can encode into the endCursor of
	// this request.
	newCursors := map[int]string{}

	var allLocations []codeintelapi.ResolvedLocation
	for i := range r.uploads {
		adjustedPath, adjustedLine, adjustedCharacter, ok, err := r.positionAdjuster.AdjustPosition(ctx, r.uploads[i].Commit, line, character)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			continue
		}

		rawCursor := ""
		if cursor, ok := cursors[r.uploads[i].ID]; ok {
			rawCursor = cursor
		} else if len(cursors) != 0 {
			// Result set is exhausted or newer than the first page
			// of results. Skip anything from this upload as it will
			// have duplicate results, or it will be out of order.
			continue
		}

		cursor, err := codeintelapi.DecodeOrCreateCursor(adjustedPath, adjustedLine, adjustedCharacter, r.uploads[i].ID, rawCursor, r.store, r.bundleManagerClient)
		if err != nil {
			return nil, "", err
		}

		locations, newCursor, hasNewCursor, err := r.codeIntelAPI.References(
			ctx,
			r.repositoryID,
			r.commit,
			limit,
			cursor,
		)
		if err != nil {
			return nil, "", err
		}

		cx := ""
		if hasNewCursor {
			cx = codeintelapi.EncodeCursor(newCursor)
		}

		allLocations = append(allLocations, locations...)

		if cx != "" {
			newCursors[r.uploads[i].ID] = cx
		}
	}

	endCursor, err := makeCursor(newCursors)
	if err != nil {
		return nil, "", err
	}

	adjustedLocations := make([]AdjustedLocation, 0, len(allLocations))
	for i := range allLocations {
		adjustedCommit := allLocations[i].Dump.Commit
		adjustedRange := convertRange(allLocations[i].Range)
		if allLocations[i].Dump.RepositoryID == r.repositoryID {
			var err error
			adjustedCommit, adjustedRange, err = r.positionAdjuster.AdjustLocation(ctx, allLocations[i].Dump.Commit, allLocations[i].Path, allLocations[i].Range)
			if err != nil {
				return nil, "", err
			}
		}

		adjustedLocations = append(adjustedLocations, AdjustedLocation{
			Dump:           allLocations[i].Dump,
			Path:           allLocations[i].Path,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return adjustedLocations, endCursor, nil
}

func (r *QueryResolver) Hover(ctx context.Context, line, character int) (string, lsp.Range, bool, error) {
	for i := range r.uploads {
		adjustedPath, adjustedLine, adjustedCharacter, ok, err := r.positionAdjuster.AdjustPosition(ctx, r.uploads[i].Commit, line, character)
		if err != nil {
			return "", lsp.Range{}, false, err
		}
		if !ok {
			continue
		}

		text, rn, exists, err := r.codeIntelAPI.Hover(ctx, adjustedPath, adjustedLine, adjustedCharacter, r.uploads[i].ID)
		if err != nil || !exists {
			return "", lsp.Range{}, false, err
		}

		lspRange := convertRange(rn)

		if text == "" {
			continue
		}

		adjustedRange, ok, err := r.positionAdjuster.AdjustRange(ctx, r.uploads[i].Commit, lspRange)
		if err != nil {
			return "", lsp.Range{}, false, err
		}
		if !ok {
			// Failed to adjust range. This _might_ happen in cases where the LSIF range
			// spans multiple lines which intersect a diff; the hover position on an earlier
			// line may not be edited, but the ending line of the expression may have been
			// edited or removed. This is rare and unfortunate, and we'll skip the result
			// in this case because we have low confidence that it will be rendered correctly.
			continue
		}

		return text, adjustedRange, true, nil
	}

	return "", lsp.Range{}, false, nil
}

func (r *QueryResolver) Diagnostics(ctx context.Context, limit int) ([]AdjustedDiagnostic, int, error) {
	totalCount := 0
	var allDiagnostics []codeintelapi.ResolvedDiagnostic
	for i := range r.uploads {
		adjustedPath, ok, err := r.positionAdjuster.AdjustPath(ctx, r.uploads[i].Commit)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			continue
		}

		l := limit - len(allDiagnostics)
		if l < 0 {
			l = 0
		}

		diagnostics, count, err := r.codeIntelAPI.Diagnostics(ctx, adjustedPath, r.uploads[i].ID, l, 0)
		if err != nil {
			return nil, 0, err
		}

		totalCount += count
		allDiagnostics = append(allDiagnostics, diagnostics...)
	}

	adjustedDiagnostics := make([]AdjustedDiagnostic, 0, len(allDiagnostics))
	for i := range allDiagnostics {
		clientRange := client.Range{
			Start: client.Position{Line: allDiagnostics[i].Diagnostic.StartLine, Character: allDiagnostics[i].Diagnostic.EndLine},
			End:   client.Position{Line: allDiagnostics[i].Diagnostic.StartCharacter, Character: allDiagnostics[i].Diagnostic.EndCharacter},
		}

		adjustedCommit := allDiagnostics[i].Dump.Commit
		adjustedRange := convertRange(clientRange)
		if allDiagnostics[i].Dump.RepositoryID == r.repositoryID {
			var err error
			adjustedCommit, adjustedRange, err = r.positionAdjuster.AdjustLocation(ctx, allDiagnostics[i].Dump.Commit, allDiagnostics[i].Diagnostic.Path, clientRange)
			if err != nil {
				return nil, 0, err
			}
		}

		adjustedDiagnostics = append(adjustedDiagnostics, AdjustedDiagnostic{
			Diagnostic:     allDiagnostics[i].Diagnostic,
			Dump:           allDiagnostics[i].Dump,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return adjustedDiagnostics, totalCount, nil
}

// readCursor decodes a cursor into a map from upload ids to URLs that
// serves the next page of results.
func readCursor(after string) (map[int]string, error) {
	if after == "" {
		return nil, nil
	}

	var cursors map[int]string
	if err := json.Unmarshal([]byte(after), &cursors); err != nil {
		return nil, err
	}
	return cursors, nil
}

// makeCursor encodes a map from upload ids to URLs that serves the next
// page of results into a single string that can be sent back for use in
// cursor pagination.
func makeCursor(cursors map[int]string) (string, error) {
	if len(cursors) == 0 {
		return "", nil
	}

	encoded, err := json.Marshal(cursors)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
