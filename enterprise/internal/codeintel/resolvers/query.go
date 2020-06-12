package resolvers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/sourcegraph/go-lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	codeintelapi "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type lsifQueryResolver struct {
	resolver *realLsifQueryResolver
}

var _ gql.GitBlobLSIFDataResolver = &lsifQueryResolver{}

func (r *lsifQueryResolver) ToGitTreeLSIFData() (gql.GitTreeLSIFDataResolver, bool) {
	return r, true
}

func (r *lsifQueryResolver) ToGitBlobLSIFData() (gql.GitBlobLSIFDataResolver, bool) {
	return r, true
}

func (r *lsifQueryResolver) Definitions(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	locations, err := r.resolver.Definitions(ctx, args)
	if err != nil {
		return nil, err
	}

	return &locationConnectionResolver{locations: locations}, nil
}

func (r *lsifQueryResolver) References(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) (gql.LocationConnectionResolver, error) {
	locations, cursor, err := r.resolver.References(ctx, args)
	if err != nil {
		return nil, err
	}

	return &locationConnectionResolver{locations: locations, endCursor: cursor}, nil
}

func (r *lsifQueryResolver) Hover(ctx context.Context, args *gql.LSIFQueryPositionArgs) (gql.HoverResolver, error) {
	text, lspRange, exists, err := r.resolver.Hover(ctx, args)
	if err != nil || !exists {
		return nil, err
	}

	return &hoverResolver{text: text, lspRange: lspRange}, nil
}

func (r *lsifQueryResolver) Diagnostics(ctx context.Context, args *gql.LSIFDiagnosticsArgs) (gql.DiagnosticConnectionResolver, error) {
	diagnostics, totalCount, err := r.resolver.Diagnostics(ctx, args)
	if err != nil {
		return nil, err
	}

	return &diagnosticConnectionResolver{totalCount: totalCount, diagnostics: diagnostics}, nil
}

//
//

type realLsifQueryResolver struct {
	store               store.Store
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI
	repo                *types.Repo
	commit              api.CommitID
	path                string
	uploads             []store.Dump
}

func (r *realLsifQueryResolver) Definitions(ctx context.Context, args *gql.LSIFQueryPositionArgs) ([]AdjustedLocation, error) {
	for _, upload := range r.uploads {
		// TODO(efritz) - we should also detect renames/copies on position adjustment
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, args.Line, args.Character)
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
					location:       l,
					adjustedCommit: adjustedCommit,
					adjustedRange:  adjustedRange,
				})
			}

			return adjustedLocations, nil
		}
	}

	return nil, nil
}

func (r *realLsifQueryResolver) References(ctx context.Context, args *gql.LSIFPagedQueryPositionArgs) ([]AdjustedLocation, string, error) {
	// Decode a map of upload ids to the next url that serves
	// the new page of results. This may not include an entry
	// for every upload if their result sets have already been
	// exhausted.
	nextURLs, err := readCursor(args.After)
	if err != nil {
		return nil, "", err
	}

	// We need to maintain a symmetric map for the next page
	// of results that we can encode into the endCursor of
	// this request.
	newCursors := map[int]string{}

	var allLocations []codeintelapi.ResolvedLocation
	for _, upload := range r.uploads {
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, args.Line, args.Character)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			continue
		}

		limit := DefaultReferencesPageSize
		if args.First != nil {
			limit = int(*args.First)
		}
		if limit <= 0 {
			// TODO(efritz) - check on defs too
			return nil, "", errors.New("illegal limit")
		}

		rawCursor := ""
		if nextURL, ok := nextURLs[upload.ID]; ok {
			rawCursor = nextURL
		} else if len(nextURLs) != 0 {
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
			location:       l,
			adjustedCommit: adjustedCommit,
			adjustedRange:  adjustedRange,
		})
	}

	return adjustedLocations, endCursor, nil
}

func (r *realLsifQueryResolver) Hover(ctx context.Context, args *gql.LSIFQueryPositionArgs) (string, lsp.Range, bool, error) {
	for _, upload := range r.uploads {
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, args.Line, args.Character)
		if err != nil {
			return "", lsp.Range{}, false, err
		}
		if !ok {
			continue
		}

		// TODO(efritz) - codeintelapi should just return an lsp.Hover
		text, rn, exists, err := r.codeIntelAPI.Hover(ctx, r.path, adjustedPosition.Line, adjustedPosition.Character, int(upload.ID))
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

func (r *realLsifQueryResolver) Diagnostics(ctx context.Context, args *gql.LSIFDiagnosticsArgs) ([]AdjustedDiagnostic, int, error) {
	limit := DefaultDiagnosticsPageSize
	if args.First != nil {
		limit = int(*args.First)
	}
	if limit <= 0 {
		return nil, 0, errors.New("illegal limit")
	}

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
			diagnostic:     d,
			adjustedCommit: adjustedCommit,
			adjustedRange:  adjustedRange,
		})
	}

	return adjustedDiagnostics, totalCount, nil
}

// adjustPosition adjusts the position denoted by `line` and `character` in the requested commit into an
// LSP position in the upload commit. This method returns nil if no equivalent position is found.
func (r *realLsifQueryResolver) adjustPosition(ctx context.Context, uploadCommit string, line, character int32) (lsp.Position, bool, error) {
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
func (r *realLsifQueryResolver) adjustLocation(ctx context.Context, location codeintelapi.ResolvedLocation) (string, lsp.Range, error) {
	return adjustLocation(ctx, location.Dump.RepositoryID, location.Dump.Commit, location.Path, location.Range, r.repo, r.commit)
}

// adjustPosition adjusts the given range in the upload commit into an equivalent range in the requested
// commit. This method returns nil if there is not an equivalent position for both endpoints of the range.
func (r *realLsifQueryResolver) adjustRange(ctx context.Context, uploadCommit string, lspRange lsp.Range) (lsp.Range, bool, error) {
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
func readCursor(after *string) (map[int]string, error) {
	if after == nil {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*after)
	if err != nil {
		return nil, err
	}

	var cursors map[int]string
	if err := json.Unmarshal(decoded, &cursors); err != nil {
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
	return base64.StdEncoding.EncodeToString(encoded), nil
}
