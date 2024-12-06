package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"log"
	"strconv"
	"strings"
)

func createPullRequest(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	sourceBranch string, targetBranch string, title string, body string) map[string]string {
	pr, _, err := client.PullRequests.Create(*ctx, repositoryOwner, repositoryName,
		&github.NewPullRequest{Title: &title, Body: &body, Head: &sourceBranch, Base: &targetBranch})
	if err != nil {
		ghErr, _ := err.(*github.ErrorResponse)
		fmt.Println(err)
		if strings.HasPrefix(ghErr.Errors[0].Message, "A pull request already exists") {
			fmt.Println("Pull Request with requested head and base already exists. Getting it's Number.")
			options := &github.PullRequestListOptions{
				Head: repositoryOwner + ":" + sourceBranch,
				Base: targetBranch,
			}
			prs, _, err := client.PullRequests.List(*ctx, repositoryOwner, repositoryName, options)
			failOnErr(err)
			pr = prs[0]
		} else {
			log.Fatal(err)
		}
	}
	fields := map[string]string{"PR_NUMBER": strconv.Itoa(*pr.Number)}
	return fields
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
