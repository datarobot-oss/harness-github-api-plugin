package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/google/go-github/v62/github"
)

// Reactions accepted by the GitHub API for issue and pull request comments.
var allowedReactions = []string{"+1", "-1", "laugh", "confused", "heart", "hooray", "rocket", "eyes"}

// commentMarkerTag returns the hidden HTML comment used to find a comment again on later runs.
// GitHub does not render it, so it never shows up in the PR conversation.
func commentMarkerTag(marker string) string {
	return fmt.Sprintf("<!-- harness-github-api-plugin:%s -->", marker)
}

// upsertPrComment creates a comment on a pull request. When marker is set, it first looks for a
// comment the token owner posted earlier with the same marker and edits that one instead, so a
// pipeline can keep a single comment per PR up to date across its stages and runs.
func upsertPrComment(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	prNumberString string, body string, marker string) map[string]string {
	prNumber, err := strconv.Atoi(prNumberString)
	failOnErr(err)

	if marker != "" {
		tag := commentMarkerTag(marker)
		body = body + "\n\n" + tag

		existing := findMarkedComment(client, ctx, repositoryName, repositoryOwner, prNumber, tag)
		if existing != nil {
			fmt.Printf("Updating comment %d (marker %q)\n", existing.GetID(), marker)
			comment, _, err := client.Issues.EditComment(*ctx, repositoryOwner, repositoryName, existing.GetID(),
				&github.IssueComment{Body: &body})
			failOnErr(err)
			return commentFields(comment, "updated")
		}
	}

	fmt.Printf("Creating comment on pull request %d\n", prNumber)
	comment, _, err := client.Issues.CreateComment(*ctx, repositoryOwner, repositoryName, prNumber,
		&github.IssueComment{Body: &body})
	failOnErr(err)
	return commentFields(comment, "created")
}

// findMarkedComment returns the most recent comment on the PR that carries tag and was written by the
// token owner, or nil. Comments by other users are ignored: the token usually cannot edit them, and
// someone quoting the bot's comment must not hijack it.
func findMarkedComment(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	prNumber int, tag string) *github.IssueComment {
	me, _, err := client.Users.Get(*ctx, "")
	failOnErr(err)

	var found *github.IssueComment
	opts := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		comments, resp, err := client.Issues.ListComments(*ctx, repositoryOwner, repositoryName, prNumber, opts)
		failOnErr(err)
		for _, c := range comments {
			if c.GetUser().GetLogin() == me.GetLogin() && strings.Contains(c.GetBody(), tag) {
				found = c
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return found
}

func commentFields(comment *github.IssueComment, action string) map[string]string {
	return map[string]string{
		"COMMENT_ID":     strconv.FormatInt(comment.GetID(), 10),
		"COMMENT_URL":    comment.GetHTMLURL(),
		"COMMENT_ACTION": action,
	}
}

// addReaction adds an emoji reaction to an issue or pull request comment, e.g. to acknowledge the
// comment that triggered a pipeline. GitHub returns the existing reaction if it is already there.
func addReaction(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	commentIdString string, reaction string) {
	if !slices.Contains(allowedReactions, reaction) {
		log.Fatalf("Unsupported reaction %q, use one of: %s", reaction, strings.Join(allowedReactions, ", "))
	}
	commentId, err := strconv.ParseInt(commentIdString, 10, 64)
	failOnErr(err)
	_, _, err = client.Reactions.CreateIssueCommentReaction(*ctx, repositoryOwner, repositoryName, commentId, reaction)
	failOnErr(err)
}

// getUserPermission reports a user's permission on the repository, so a pipeline started by a PR
// comment can check that the commenter is allowed to run it. A user GitHub does not know is reported
// as "none" instead of failing the step, so the caller decides what to do.
func getUserPermission(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	username string) map[string]string {
	permission := "none"
	level, resp, err := client.Repositories.GetPermissionLevel(*ctx, repositoryOwner, repositoryName, username)
	if err != nil {
		if resp == nil || resp.StatusCode != 404 {
			log.Fatal(err)
		}
		fmt.Printf("User %s not found, reporting permission none\n", username)
	} else {
		permission = level.GetPermission()
	}
	return map[string]string{
		"PERMISSION":       permission,
		"HAS_WRITE_ACCESS": strconv.FormatBool(permission == "admin" || permission == "write"),
	}
}
