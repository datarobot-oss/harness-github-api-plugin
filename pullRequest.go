package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
	"log"
	"slices"
	"strconv"
	"strings"
)

func createPullRequest(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	sourceBranch string, targetBranch string, title string, body string, assignees string, labels string) map[string]string {
	pr, _, err := client.PullRequests.Create(*ctx, repositoryOwner, repositoryName,
		&github.NewPullRequest{Title: &title, Body: &body, Head: &sourceBranch, Base: &targetBranch})
	if err != nil {
		ghErr, _ := err.(*github.ErrorResponse)
		if strings.HasPrefix(ghErr.Errors[0].Message, "A pull request already exists") {
			fmt.Println("Pull Request with requested head and base already exists. Updating Title and Body. Getting PR details.")
			options := &github.PullRequestListOptions{
				Head: repositoryOwner + ":" + sourceBranch,
				Base: targetBranch,
			}
			prs, _, err := client.PullRequests.List(*ctx, repositoryOwner, repositoryName, options)
			failOnErr(err)
			pr = prs[0]
			pr.Title = &title
			pr.Body = &body
			pr, _, err = client.PullRequests.Edit(*ctx, repositoryOwner, repositoryName, *pr.Number, pr)
			failOnErr(err)
		} else {
			log.Fatal(err)
		}
	}
	if assignees != "" {
		fmt.Println("Adding assignees: " + assignees)
		addPullRequestAssignees(client, ctx, repositoryName, repositoryOwner, *pr.Number, assignees)
	}
	if labels != "" && labels != "null" {
		fmt.Println("Adding labels: " + labels)
		addPullRequestLabels(client, ctx, repositoryName, repositoryOwner, strconv.Itoa(*pr.Number), labels)
	}
	fields := map[string]string{
		"PR_NUMBER":   strconv.Itoa(*pr.Number),
		"PR_URL":      pr.GetHTMLURL(),
		"PR_TITLE":    pr.GetTitle(),
		"PR_HEAD":     pr.GetHead().GetRef(),
		"PR_HEAD_SHA": pr.GetHead().GetSHA(),
		"PR_BASE":     pr.GetBase().GetRef(),
	}
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

func addPullRequestLabels(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, pullRequestNumber string, labels string) {
	number, _ := strconv.Atoi(pullRequestNumber)
	_, _, err := client.Issues.AddLabelsToIssue(*ctx, repositoryOwner, repositoryName, number, splitList(labels))
	failOnErr(err)
}

// addPullRequestAssignees adds the users to the PR's assignees. GitHub silently skips users who cannot
// be assigned (for example without access to the repository), so they are reported as a warning.
func addPullRequestAssignees(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	pullRequestNumber int, assignees string) {
	logins := splitList(assignees)
	issue, _, err := client.Issues.AddAssignees(*ctx, repositoryOwner, repositoryName, pullRequestNumber, logins)
	failOnErr(err)
	for _, login := range logins {
		if !slices.ContainsFunc(issue.Assignees, func(user *github.User) bool {
			return strings.EqualFold(user.GetLogin(), login)
		}) {
			fmt.Printf("Warning: could not assign %s to pull request %d\n", login, pullRequestNumber)
		}
	}
}
