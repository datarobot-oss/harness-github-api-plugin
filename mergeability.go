package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v62/github"
)

// GitHub computes a pull request's test merge commit (refs/pull/<N>/merge) in the background after
// every push to the head or base branch. Until it is done, the API returns "mergeable": null, and
// "merge_commit_sha" can still name the merge of an earlier head. waitForMergeCommit polls until the
// test merge is computed for the expected head, so a caller can build exactly what would be merged.

const (
	defaultMergeableTimeout  = 120 * time.Second
	defaultMergeablePollTime = 5 * time.Second
)

// prReader is the part of the GitHub API waitForMergeCommit needs; tests use a fake.
type prReader interface {
	getPullRequest(ctx context.Context, number int) (*github.PullRequest, error)
	getCommitParents(ctx context.Context, sha string) ([]string, error)
}

type githubPrReader struct {
	client *github.Client
	owner  string
	repo   string
}

func (r githubPrReader) getPullRequest(ctx context.Context, number int) (*github.PullRequest, error) {
	pr, _, err := r.client.PullRequests.Get(ctx, r.owner, r.repo, number)
	return pr, err
}

func (r githubPrReader) getCommitParents(ctx context.Context, sha string) ([]string, error) {
	commit, _, err := r.client.Git.GetCommit(ctx, r.owner, r.repo, sha)
	if err != nil {
		return nil, err
	}
	var parents []string
	for _, p := range commit.Parents {
		parents = append(parents, p.GetSHA())
	}
	return parents, nil
}

type mergeWaitOptions struct {
	// expectedHeadSha, when set, fails the wait if the pull request head is another commit
	// (the branch moved after the caller resolved it).
	expectedHeadSha string
	timeout         time.Duration
	pollInterval    time.Duration
	sleep           func(time.Duration)
}

// waitForMergeCommit returns the pull request once GitHub reports it mergeable and its
// merge_commit_sha is a merge of the current head. It fails when the pull request is not open,
// has merge conflicts, its head is not expectedHeadSha, or the merge is not computed in time.
func waitForMergeCommit(ctx context.Context, r prReader, number int, opts mergeWaitOptions) (*github.PullRequest, error) {
	if opts.timeout <= 0 {
		opts.timeout = defaultMergeableTimeout
	}
	if opts.pollInterval <= 0 {
		opts.pollInterval = defaultMergeablePollTime
	}
	if opts.sleep == nil {
		opts.sleep = time.Sleep
	}

	var waited time.Duration
	for {
		pr, err := r.getPullRequest(ctx, number)
		if err != nil {
			return nil, err
		}
		head := pr.GetHead().GetSHA()
		if pr.GetState() != "open" {
			return nil, fmt.Errorf("pull request #%d is %s; a merge commit exists only for open pull requests", number, pr.GetState())
		}
		if opts.expectedHeadSha != "" && head != opts.expectedHeadSha {
			return nil, fmt.Errorf("pull request #%d head is %s, expected %s: the branch moved, a newer run should follow", number, head, opts.expectedHeadSha)
		}

		pending := ""
		switch {
		case pr.Mergeable == nil:
			pending = "GitHub has not computed mergeability yet"
		case !pr.GetMergeable():
			return nil, fmt.Errorf("pull request #%d has merge conflicts with %s (mergeable_state %q): rebase it or merge %s into it",
				number, pr.GetBase().GetRef(), pr.GetMergeableState(), pr.GetBase().GetRef())
		case pr.GetMergeCommitSHA() == "":
			pending = "no merge commit yet"
		default:
			parents, err := r.getCommitParents(ctx, pr.GetMergeCommitSHA())
			if err != nil {
				return nil, err
			}
			// A test merge commit has the base branch as first parent and the head as second.
			if len(parents) == 2 && parents[1] == head {
				return pr, nil
			}
			pending = fmt.Sprintf("merge commit %s is not a merge of head %s yet (parents %v)", pr.GetMergeCommitSHA(), head, parents)
		}

		if waited >= opts.timeout {
			return nil, fmt.Errorf("pull request #%d: %s after %s", number, pending, opts.timeout)
		}
		fmt.Printf("Waiting for pull request #%d merge commit: %s\n", number, pending)
		opts.sleep(opts.pollInterval)
		waited += opts.pollInterval
	}
}

// mergeWaitOptionsFromSettings reads the optional getPrDetails settings. It returns enabled=false
// when PR_WAIT_MERGEABLE is not "true".
func mergeWaitOptionsFromSettings(waitMergeable, expectedHeadSha, timeoutSeconds string) (mergeWaitOptions, bool, error) {
	if waitMergeable != "true" {
		return mergeWaitOptions{}, false, nil
	}
	opts := mergeWaitOptions{expectedHeadSha: expectedHeadSha}
	if timeoutSeconds != "" {
		seconds, err := strconv.Atoi(timeoutSeconds)
		if err != nil || seconds <= 0 {
			return opts, true, fmt.Errorf("PR_MERGEABLE_TIMEOUT must be a positive number of seconds, got %q", timeoutSeconds)
		}
		opts.timeout = time.Duration(seconds) * time.Second
	}
	return opts, true, nil
}
