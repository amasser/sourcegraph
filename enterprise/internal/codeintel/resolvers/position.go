package resolvers

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// TODO - document, test

type PositionAdjuster interface {
	AdjustPath(ctx context.Context, commit, path string, reverse bool) (string, bool, error)
	AdjustPosition(ctx context.Context, commit, path string, px bundles.Position, reverse bool) (string, bundles.Position, bool, error)
	AdjustRange(ctx context.Context, commit, path string, rx bundles.Range, reverse bool) (string, bundles.Range, bool, error)
}

type realPositionAdjuster struct {
	repo            *types.Repo
	requestedCommit string
}

func NewPositionAdjuster(repo *types.Repo, requestedCommit string) PositionAdjuster {
	return &realPositionAdjuster{
		repo:            repo,
		requestedCommit: requestedCommit,
	}
}

func (p *realPositionAdjuster) AdjustPath(ctx context.Context, commit, path string, reverse bool) (string, bool, error) {
	return path, true, nil
}

func (p *realPositionAdjuster) AdjustPosition(ctx context.Context, commit, path string, px bundles.Position, reverse bool) (string, bundles.Position, bool, error) {
	hunks, err := readHunks(ctx, p.repo, p.requestedCommit, commit, path, reverse)
	if err != nil {
		return "", bundles.Position{}, false, err
	}

	adjusted, ok := adjustPosition(hunks, px)
	return path, adjusted, ok, nil
}

func (p *realPositionAdjuster) AdjustRange(ctx context.Context, commit, path string, rx bundles.Range, reverse bool) (string, bundles.Range, bool, error) {
	hunks, err := readHunks(ctx, p.repo, p.requestedCommit, commit, path, reverse)
	if err != nil {
		return "", bundles.Range{}, false, err
	}

	adjusted, ok := adjustRange(hunks, rx)
	return path, adjusted, ok, nil
}

func readHunks(ctx context.Context, repo *types.Repo, sourceCommit, targetCommit, path string, reverse bool) ([]*diff.Hunk, error) {
	if sourceCommit == targetCommit {
		return nil, nil
	}

	cachedRepo, err := backend.CachedGitRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	if reverse {
		sourceCommit, targetCommit = targetCommit, sourceCommit
	}
	reader, err := git.ExecReader(ctx, *cachedRepo, []string{"diff", sourceCommit, targetCommit, "--", path}) // TODO - cache this
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	output, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}

	diff, err := diff.NewFileDiffReader(bytes.NewReader(output)).Read()
	if err != nil {
		return nil, err
	}
	return diff.Hunks, nil
}

func adjustRange(hunks []*diff.Hunk, r bundles.Range) (bundles.Range, bool) {
	start, ok := adjustPosition(hunks, r.Start)
	if !ok {
		return bundles.Range{}, false
	}

	end, ok := adjustPosition(hunks, r.End)
	if !ok {
		return bundles.Range{}, false
	}

	return bundles.Range{Start: start, End: end}, true
}

// adjustPosition transforms the given position in the source commit to a position in the target
// commit. This method returns second boolean value indicating that the adjustment succeeded. If
// that particular line does not exist or has been edited in between the source and target commit,
// then adjustment will fail. The given position is assumed to be zero-indexed.
func adjustPosition(hunks []*diff.Hunk, pos bundles.Position) (bundles.Position, bool) {
	// Find the index of the first hunk that starts after the target line and use the
	// previous hunk (if it exists) as the point of reference in `adjustPositionFromHunk`.
	// Note: LSP Positions are zero-indexed; the output of git diff is one-indexed.

	i := 0
	for i < len(hunks) && int(hunks[i].OrigStartLine) <= pos.Line+1 {
		i++
	}

	if i == 0 {
		// Trivial case, no changes before this line
		return pos, true
	}

	hunk := hunks[i-1]
	// 	return adjustPositionFromHunk(hunk, pos)
	// }

	// // adjustPositionFromHunk transforms the given position in the original file into a position
	// // in the new file according to the given git diff hunk. This parameter is expected to be the
	// // last such hunk in the diff between the original and the new file that does not begin after
	// // the given position in the original file.
	// func adjustPositionFromHunk(hunk *diff.Hunk, pos bundles.Position) (bundles.Position, bool) {
	// LSP Positions are zero-indexed; the output of git diff is one-indexed
	line := pos.Line + 1

	if line >= int(hunk.OrigStartLine+hunk.OrigLines) {
		// Hunk ends before this line, so we can simply adjust the line offset by the
		// relative difference between the line offsets in each file after this hunk.
		origEnd := int(hunk.OrigStartLine + hunk.OrigLines)
		newEnd := int(hunk.NewStartLine + hunk.NewLines)
		relativeDifference := newEnd - origEnd

		return bundles.Position{
			Line:      line + relativeDifference - 1,
			Character: pos.Character,
		}, true
	}

	// Create two fingers pointing at the first line of this hunk in each file. Then,
	// bump each of these cursors for every line in hunk body that is attributed
	// to the corresponding file.

	origFileOffset := int(hunk.OrigStartLine)
	newFileOffset := int(hunk.NewStartLine)

	for _, bodyLine := range strings.Split(string(hunk.Body), "\n") {
		// Bump original file offset unless it's an addition in the new file
		added := strings.HasPrefix(bodyLine, "+")
		if !added {
			origFileOffset++
		}

		// Bump new file offset unless it's a deletion of a line from the new file
		removed := strings.HasPrefix(bodyLine, "-")
		if !removed {
			newFileOffset++
		}

		// Keep skipping ahead in the original file until we hit our target line
		if origFileOffset-1 < line {
			continue
		}

		// This line exists in both files
		if !added && !removed {
			return bundles.Position{
				Line:      newFileOffset - 2,
				Character: pos.Character,
			}, true
		}

		// Fail the position adjustment. This particular line was either
		//   (1) edited;
		//   (2) removed in which case we can't point to it; or
		//   (3) added, in which case it hasn't been indexed.
		//
		// In all cases we don't want to return any results here as we
		// don't have enough information to give a precise result matching
		// the current query text.

		return bundles.Position{}, false
	}

	// This should never happen unless the git diff content is malformed. We know
	// the target line occurs within the hunk, but iteration of the hunk's body did
	// not contain enough lines attributed to the original file.
	panic("Malformed hunk body")
}
