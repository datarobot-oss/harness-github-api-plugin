package main

import (
	"context"
	"github.com/google/go-github/v62/github"
)

func setStatusCheck(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	sha string, statusContext string, status string, url string, description string) {
	_, _, err := client.Repositories.CreateStatus(*ctx, repositoryOwner, repositoryName, sha, &github.RepoStatus{
		State:       &status,
		Description: &description,
		Context:     &statusContext,
		TargetURL:   &url,
	})

	failOnErr(err)
}
