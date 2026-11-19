package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v62/github"
)

const (
	baseSha     = "1111111111111111111111111111111111111111"
	headSha     = "2222222222222222222222222222222222222222"
	oldHeadSha  = "3333333333333333333333333333333333333333"
	mergeSha    = "4444444444444444444444444444444444444444"
	oldMergeSha = "5555555555555555555555555555555555555555"
)

// fakePrReader returns the pull requests in prs one per call (the last one repeats) and the
// commit parents in parents.
type fakePrReader struct {
	prs     []*github.PullRequest
	parents map[string][]string
	calls   int
}

func (f *fakePrReader) getPullRequest(_ context.Context, _ int) (*github.PullRequest, error) {
	pr := f.prs[min(f.calls, len(f.prs)-1)]
	f.calls++
	return pr, nil
}

func (f *fakePrReader) getCommitParents(_ context.Context, sha string) ([]string, error) {
	return f.parents[sha], nil
}

func pr(state string, head string, mergeable *bool, mergeSha string) *github.PullRequest {
	return &github.PullRequest{
		State:          github.String(state),
		Head:           &github.PullRequestBranch{SHA: github.String(head)},
		Base:           &github.PullRequestBranch{SHA: github.String(baseSha), Ref: github.String("main")},
		Mergeable:      mergeable,
		MergeableState: github.String("dirty"),
		MergeCommitSHA: github.String(mergeSha),
	}
}

func testOptions(expectedHead string) (mergeWaitOptions, *time.Duration) {
	slept := new(time.Duration)
	return mergeWaitOptions{
		expectedHeadSha: expectedHead,
		timeout:         20 * time.Second,
		pollInterval:    5 * time.Second,
		sleep:           func(d time.Duration) { *slept += d },
	}, slept
}

var merges = map[string][]string{
	mergeSha:    {baseSha, headSha},
	oldMergeSha: {baseSha, oldHeadSha},
}

func TestWaitForMergeCommitReady(t *testing.T) {
	r := &fakePrReader{prs: []*github.PullRequest{pr("open", headSha, github.Bool(true), mergeSha)}, parents: merges}
	opts, slept := testOptions(headSha)
	got, err := waitForMergeCommit(context.Background(), r, 7, opts)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetMergeCommitSHA() != mergeSha || *slept != 0 {
		t.Fatalf("got merge %s after %s, want %s at once", got.GetMergeCommitSHA(), *slept, mergeSha)
	}
}

func TestWaitForMergeCommitWaitsForMergeability(t *testing.T) {
	r := &fakePrReader{prs: []*github.PullRequest{
		pr("open", headSha, nil, oldMergeSha),               // not computed yet
		pr("open", headSha, github.Bool(true), oldMergeSha), // merge of the previous head
		pr("open", headSha, github.Bool(true), mergeSha),
	}, parents: merges}
	opts, slept := testOptions(headSha)
	got, err := waitForMergeCommit(context.Background(), r, 7, opts)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetMergeCommitSHA() != mergeSha || *slept != 10*time.Second {
		t.Fatalf("got merge %s after %s, want %s after 10s", got.GetMergeCommitSHA(), *slept, mergeSha)
	}
}

func TestWaitForMergeCommitConflict(t *testing.T) {
	r := &fakePrReader{prs: []*github.PullRequest{pr("open", headSha, github.Bool(false), "")}, parents: merges}
	opts, slept := testOptions("")
	_, err := waitForMergeCommit(context.Background(), r, 7, opts)
	if err == nil || !strings.Contains(err.Error(), "merge conflicts with main") || *slept != 0 {
		t.Fatalf("got %v after %s, want an immediate conflict error", err, *slept)
	}
}

func TestWaitForMergeCommitHeadMoved(t *testing.T) {
	r := &fakePrReader{prs: []*github.PullRequest{pr("open", headSha, github.Bool(true), mergeSha)}, parents: merges}
	opts, _ := testOptions(oldHeadSha)
	_, err := waitForMergeCommit(context.Background(), r, 7, opts)
	if err == nil || !strings.Contains(err.Error(), "the branch moved") {
		t.Fatalf("got %v, want a head mismatch error", err)
	}
}

func TestWaitForMergeCommitNotOpen(t *testing.T) {
	r := &fakePrReader{prs: []*github.PullRequest{pr("closed", headSha, nil, mergeSha)}, parents: merges}
	opts, _ := testOptions("")
	_, err := waitForMergeCommit(context.Background(), r, 7, opts)
	if err == nil || !strings.Contains(err.Error(), "is closed") {
		t.Fatalf("got %v, want a not-open error", err)
	}
}

func TestWaitForMergeCommitTimeout(t *testing.T) {
	r := &fakePrReader{prs: []*github.PullRequest{pr("open", headSha, nil, "")}, parents: merges}
	opts, slept := testOptions("")
	_, err := waitForMergeCommit(context.Background(), r, 7, opts)
	if err == nil || !strings.Contains(err.Error(), "has not computed mergeability yet after 20s") {
		t.Fatalf("got %v, want a timeout error", err)
	}
	if *slept != 20*time.Second || r.calls != 5 {
		t.Fatalf("slept %s over %d calls, want 20s over 5 calls", *slept, r.calls)
	}
}

func TestMergeWaitOptionsFromSettings(t *testing.T) {
	if _, enabled, err := mergeWaitOptionsFromSettings("", headSha, ""); enabled || err != nil {
		t.Fatalf("unset PR_WAIT_MERGEABLE: enabled=%v err=%v, want disabled", enabled, err)
	}
	opts, enabled, err := mergeWaitOptionsFromSettings("true", headSha, "300")
	if !enabled || err != nil || opts.timeout != 300*time.Second || opts.expectedHeadSha != headSha {
		t.Fatalf("got %+v enabled=%v err=%v", opts, enabled, err)
	}
	if _, _, err := mergeWaitOptionsFromSettings("true", "", "soon"); err == nil {
		t.Fatal("want an error for a non-numeric timeout")
	}
}
