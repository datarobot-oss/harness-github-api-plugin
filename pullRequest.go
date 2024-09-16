package main

import (
	"context"
	"github.com/google/go-github/v62/github"
	"strconv"
)

func createPullRequest(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	sourceBranch string, targetBranch string, title string, body string) {
	_, _, err := client.PullRequests.Create(*ctx, repositoryOwner, repositoryName,
		&github.NewPullRequest{Title: &title, Body: &body, Head: &sourceBranch, Base: &targetBranch})
	failOnErr(err)
}

func mergePullRequest(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, pullRequestNumber string, commitMessage string) {
	number, _ := strconv.Atoi(pullRequestNumber)
	_, _, err := client.PullRequests.Merge(*ctx, repositoryOwner, repositoryName, number, commitMessage, nil)
	failOnErr(err)
}

func getPullRequest(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, prNumberString string) map[string]string {
	prNumber, err := strconv.Atoi(prNumberString)
	failOnErr(err)
	pr, _, err := client.PullRequests.Get(*ctx, repositoryOwner, repositoryName, prNumber)
	failOnErr(err)
	fields := map[string]string{
		"TITLE":            pr.GetTitle(),
		"BODY":             pr.GetBody(),
		"USER_LOGIN":       pr.GetUser().GetLogin(),
		"USER_EMAIL":       pr.GetUser().GetEmail(),
		"BASE_BRANCH_NAME": pr.GetBase().GetRef(),
		"BASE_SHA":         pr.GetBase().GetSHA(),
		"HEAD_BRANCH_NAME": pr.GetHead().GetRef(),
		"HEAD_SHA":         pr.GetHead().GetSHA(),
		"MERGE_COMMIT_SHA": pr.GetMergeCommitSHA(),
		"STATE":            pr.GetState(),
		"URL":              pr.GetURL(),
		"CREATED_AT":       pr.GetCreatedAt().String(),
		"IS_DRAFT":         strconv.FormatBool(pr.GetDraft()),
	}
	return fields
}
