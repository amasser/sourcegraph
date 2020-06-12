package mock

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type MockedGitHubChangesetSyncState struct {
	execReader     func([]string) (io.ReadCloser, error)
	mockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)
}

// GitHubChangesetSync sets up mocks such that invoking LoadChangesets() on one
// or more GitHub changesets will always return succeed, and return the same
// diff (+1, ~1, -3).
//
// UnmockGitHubChangesetSync() must called to clean up, usually via defer.
func GitHubChangesetSync(repo *protocol.RepoInfo) *MockedGitHubChangesetSyncState {
	state := &MockedGitHubChangesetSyncState{
		execReader:     git.Mocks.ExecReader,
		mockRepoLookup: repoupdater.MockRepoLookup,
	}

	repoupdater.MockRepoLookup = func(args protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error) {
		return &protocol.RepoLookupResult{
			Repo: repo,
		}, nil
	}

	git.Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
		if len(args) < 1 && args[0] != "diff" {
			if state.execReader != nil {
				return state.execReader(args)
			}
			return nil, errors.New("cannot handle non-diff command in mock ExecReader")
		}
		return ioutil.NopCloser(strings.NewReader(testDiff)), nil
	}

	return state
}

// UnmockGitHubChangesetSync resets the mocks set up by GitHubChangesetSync.
func UnmockGitHubChangesetSync(state *MockedGitHubChangesetSyncState) {
	git.Mocks.ExecReader = state.execReader
	repoupdater.MockRepoLookup = state.mockRepoLookup
}

// testDiff provides a diff that will resolve to 1 added line, 1 changed line,
// and 3 deleted lines.
const testDiff = `
diff --git a/test.py b/test.py
index 884601b..c4886d5 100644
--- a/test.py
+++ b/test.py
@@ -1,6 +1,4 @@
+# square makes a value squarer.
 def square(a):
-    """
-    square makes a value squarer.
-    """

-    return a * a
+    return pow(a, 2)

`
