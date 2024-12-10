package main

import (
	"context"
	"fmt"
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

func listStatusChecks(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, sha string) map[string]string {
	statuses, _, err := client.Repositories.ListStatuses(*ctx, repositoryOwner, repositoryName, sha, nil)
	failOnErr(err)
	var allStatuses string = ""
	for i, status := range statuses {
		var prefix string = ""
		if i > 0 {
			prefix = ", "
		}
		allStatuses += prefix + *status.Context + " " + *status.State
	}
	fmt.Println(allStatuses)
	fields := map[string]string{"STATUSES": allStatuses}
	return fields
}
