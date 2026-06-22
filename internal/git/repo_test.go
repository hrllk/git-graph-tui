package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenReturnsAbsoluteRoot(t *testing.T) {
	repo, err := Open(".")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if repo.MustRoot() == "" {
		t.Fatal("expected root to be set")
	}
	if !filepath.IsAbs(repo.MustRoot()) {
		t.Fatalf("expected absolute root, got %q", repo.MustRoot())
	}
}

func TestRunnerRejectsUnknownCommand(t *testing.T) {
	r := &Runner{}
	_, err := r.Run("not-a-real-git-command")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsNoCommits(t *testing.T) {
	err := os.ErrNotExist
	if isNoCommits(err) {
		t.Fatal("expected unrelated error to not be treated as no-commits")
	}
}

func TestFilterRemoteBranchesDropsSymbolicHead(t *testing.T) {
	got := filterRemoteBranches([]string{"origin/HEAD", "origin/main", "origin/tmp3"})
	if len(got) != 2 || got[0] != "origin/main" || got[1] != "origin/tmp3" {
		t.Fatalf("unexpected filtered remote branches: %v", got)
	}
}

func TestParseTrackingInfo(t *testing.T) {
	tests := []struct {
		input  string
		ahead  int
		behind int
	}{
		{"[ahead 1, behind 2]", 1, 2},
		{"[ahead 5]", 5, 0},
		{"[behind 3]", 0, 3},
		{"[gone]", 0, 0},
		{"", 0, 0},
	}

	for _, tc := range tests {
		a, b := parseTrackingInfo(tc.input)
		if a != tc.ahead || b != tc.behind {
			t.Errorf("parseTrackingInfo(%q) = (%d, %d); want (%d, %d)", tc.input, a, b, tc.ahead, tc.behind)
		}
	}
}
