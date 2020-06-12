package resolvers

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
	repo                *types.Repo
	commit              api.CommitID
	path                string
	uploads             []store.Dump
}

func NewQueryResolver(
	store store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	codeIntelAPI codeintelapi.CodeIntelAPI,
	repo *types.Repo,
	commit api.CommitID,
	path string,
	uploads []store.Dump,
) *QueryResolver {
	return &QueryResolver{
		store:               store,
		bundleManagerClient: bundleManagerClient,
		codeIntelAPI:        codeIntelAPI,
		repo:                repo,
		commit:              commit,
		path:                path,
		uploads:             uploads,
	}
}

func (r *QueryResolver) Definitions(ctx context.Context, line, character int) ([]AdjustedLocation, error) {
	for _, upload := range r.uploads {
		// TODO(efritz) - we should also detect renames/copies on position adjustment
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, int32(line), int32(character))
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		locations, err := r.codeIntelAPI.Definitions(ctx, r.path, adjustedPosition.Line, adjustedPosition.Character, upload.ID)
		if err != nil {
			return nil, err
		}

		if len(locations) > 0 {
			var adjustedLocations []AdjustedLocation
			for _, l := range locations {
				adjustedCommit, adjustedRange, err := r.adjustLocation(ctx, l)
				if err != nil {
					return nil, err
				}

				adjustedLocations = append(adjustedLocations, AdjustedLocation{
					Dump:           l.Dump,
					Path:           l.Path,
					AdjustedCommit: adjustedCommit,
					AdjustedRange:  adjustedRange,
				})
			}

			return adjustedLocations, nil
		}
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
	for _, upload := range r.uploads {
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, int32(line), int32(character))
		if err != nil {
			return nil, "", err
		}
		if !ok {
			continue
		}

		rawCursor := ""
		if cursor, ok := cursors[upload.ID]; ok {
			rawCursor = cursor
		} else if len(cursors) != 0 {
			// Result set is exhausted or newer than the first page
			// of results. Skip anything from this upload as it will
			// have duplicate results, or it will be out of order.
			continue
		}

		cursor, err := codeintelapi.DecodeOrCreateCursor(r.path, adjustedPosition.Line, adjustedPosition.Character, upload.ID, rawCursor, r.store, r.bundleManagerClient)
		if err != nil {
			return nil, "", err
		}

		locations, newCursor, hasNewCursor, err := r.codeIntelAPI.References(
			ctx,
			int(r.repo.ID),
			string(r.commit),
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
			newCursors[upload.ID] = cx
		}
	}

	endCursor, err := makeCursor(newCursors)
	if err != nil {
		return nil, "", err
	}

	var adjustedLocations []AdjustedLocation
	for _, l := range allLocations {
		adjustedCommit, adjustedRange, err := r.adjustLocation(ctx, l)
		if err != nil {
			return nil, "", err
		}

		adjustedLocations = append(adjustedLocations, AdjustedLocation{
			Dump:           l.Dump,
			Path:           l.Path,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return adjustedLocations, endCursor, nil
}

func (r *QueryResolver) Hover(ctx context.Context, line, character int) (string, lsp.Range, bool, error) {
	for _, upload := range r.uploads {
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, int32(line), int32(character))
		if err != nil {
			return "", lsp.Range{}, false, err
		}
		if !ok {
			continue
		}

		text, rn, exists, err := r.codeIntelAPI.Hover(ctx, r.path, adjustedPosition.Line, adjustedPosition.Character, upload.ID)
		if err != nil || !exists {
			return "", lsp.Range{}, false, err
		}

		lspRange := convertRange(rn)

		if text != "" {
			adjustedRange, ok, err := r.adjustRange(ctx, upload.Commit, lspRange)
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
	}

	return "", lsp.Range{}, false, nil
}

func (r *QueryResolver) Diagnostics(ctx context.Context, limit int) ([]AdjustedDiagnostic, int, error) {
	totalCount := 0
	var allDiagnostics []codeintelapi.ResolvedDiagnostic
	for _, upload := range r.uploads {
		l := limit - len(allDiagnostics)
		if l < 0 {
			l = 0
		}

		diagnostics, count, err := r.codeIntelAPI.Diagnostics(ctx, r.path, upload.ID, l, 0)
		if err != nil {
			return nil, 0, err
		}

		totalCount += count
		allDiagnostics = append(allDiagnostics, diagnostics...)
	}

	var adjustedDiagnostics []AdjustedDiagnostic
	for _, d := range allDiagnostics {
		clientRange := client.Range{
			Start: client.Position{Line: d.Diagnostic.StartLine, Character: d.Diagnostic.EndLine},
			End:   client.Position{Line: d.Diagnostic.StartCharacter, Character: d.Diagnostic.EndCharacter},
		}

		adjustedCommit, adjustedRange, err := adjustLocation(ctx, d.Dump.RepositoryID, d.Dump.Commit, d.Diagnostic.Path, clientRange, r.repo, r.commit)
		if err != nil {
			return nil, 0, err
		}

		adjustedDiagnostics = append(adjustedDiagnostics, AdjustedDiagnostic{
			Diagnostic:     d.Diagnostic,
			Dump:           d.Dump,
			AdjustedCommit: adjustedCommit,
			AdjustedRange:  adjustedRange,
		})
	}

	return adjustedDiagnostics, totalCount, nil
}

// adjustPosition adjusts the position denoted by `line` and `character` in the requested commit into an
// LSP position in the upload commit. This method returns nil if no equivalent position is found.
func (r *QueryResolver) adjustPosition(ctx context.Context, uploadCommit string, line, character int32) (lsp.Position, bool, error) {
	adjuster, err := newPositionAdjuster(ctx, r.repo, string(r.commit), uploadCommit, r.path)
	if err != nil {
		return lsp.Position{}, false, err
	}

	adjusted, ok := adjuster.adjustPosition(lsp.Position{Line: int(line), Character: int(character)})
	return adjusted, ok, nil
}

// adjustLocation attempts to transform the source range of location into a corresponding
// range of the same file at the user's requested commit.
//
// If location has no corresponding range at the requested commit or is located in a different
// repository, it returns the location's current commit and range without modification.
// Otherwise, it returns the user's requested commit along with the transformed range.
//
// A non-nil error means the connection resolver was unable to load the diff between
// the requested commit and location's commit.
func (r *QueryResolver) adjustLocation(ctx context.Context, location codeintelapi.ResolvedLocation) (string, lsp.Range, error) {
	return adjustLocation(ctx, location.Dump.RepositoryID, location.Dump.Commit, location.Path, location.Range, r.repo, r.commit)
}

// adjustPosition adjusts the given range in the upload commit into an equivalent range in the requested
// commit. This method returns nil if there is not an equivalent position for both endpoints of the range.
func (r *QueryResolver) adjustRange(ctx context.Context, uploadCommit string, lspRange lsp.Range) (lsp.Range, bool, error) {
	adjuster, err := newPositionAdjuster(ctx, r.repo, uploadCommit, string(r.commit), r.path)
	if err != nil {
		return lsp.Range{}, false, err
	}

	adjusted, ok := adjuster.adjustRange(lspRange)
	return adjusted, ok, nil
}

func adjustLocation(ctx context.Context, locationRepositoryID int, locationCommit, locationPath string, locationRange bundles.Range, repo *types.Repo, commit api.CommitID) (string, lsp.Range, error) {
	if api.RepoID(locationRepositoryID) != repo.ID {
		return locationCommit, convertRange(locationRange), nil
	}

	adjuster, err := newPositionAdjuster(ctx, repo, locationCommit, string(commit), locationPath)
	if err != nil {
		return "", lsp.Range{}, err
	}

	if adjustedRange, ok := adjuster.adjustRange(convertRange(locationRange)); ok {
		return string(commit), adjustedRange, nil
	}

	// Couldn't adjust range, return original result which is precise but
	// jump the user to another into another commit context on navigation.
	return locationCommit, convertRange(locationRange), nil
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

func convertRange(r bundles.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

func convertPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}
