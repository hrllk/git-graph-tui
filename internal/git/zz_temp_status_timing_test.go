package git

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestStatusTiming(t *testing.T) {
	root := os.Getenv("BENCH_REPO")
	if root == "" {
		t.Skip("BENCH_REPO not set")
	}
	repo, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	st, err := repo.Status(context.Background(), 40)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("elapsed=%s commits=%d branches=%d remotes=%d tags=%d", elapsed, len(st.GraphCommits), len(st.LocalBranches), len(st.RemoteBranches), len(st.Tags))
}
