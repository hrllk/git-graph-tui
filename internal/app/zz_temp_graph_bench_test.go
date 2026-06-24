package app

import (
	"fmt"
	"testing"

	"hrllk/git-graph-tui/internal/git"
	"hrllk/git-graph-tui/internal/state"
)

func BenchmarkGraphView5000(b *testing.B) {
	commits := make([]git.GraphCommit, 5000)
	for i := range commits {
		parents := []string{}
		if i > 0 {
			parents = []string{fmt.Sprintf("%04d", i-1)}
		}
		commits[i] = git.GraphCommit{
			Graph:       "* ",
			Hash:        fmt.Sprintf("%04d", i),
			Parents:     parents,
			RelativeAge: "0 seconds ago",
			Author:      "Test",
			Subject:     "commit",
		}
	}
	m := model{
		status:        state.New().WithBrowse(),
		activeSection: sectionGraph,
		sectionCursor: map[graphSection]int{sectionGraph: 0, sectionCurrent: 0, sectionLocal: 0, sectionRemote: 0, sectionTags: 0},
		repoStatus: git.Status{
			Branch:       "main",
			Head:         "0000",
			GraphCommits: commits,
			LocalBranches: []string{"main"},
		},
		width:  120,
		height: 80,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = m.renderGraphContent(80, 20)
	}
}
