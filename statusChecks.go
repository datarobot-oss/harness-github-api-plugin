package main

import (
	"context"
<<<<<<< HEAD
	"encoding/json"
	"github.com/google/go-github/v62/github"
	"strings"
=======
	"fmt"
	"github.com/google/go-github/v62/github"
>>>>>>> 4c8b004 (Add getStatuses command)
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
	allStatuses := ""
	allStatusesMap := map[string]string{}
	for i, status := range statuses {
		prefix := ""
		if i > 0 {
			prefix = ", "
		}
		if !strings.Contains(allStatuses, *status.Context) {
			allStatuses += prefix + *status.Context + " " + *status.State
			allStatusesMap[*status.Context] = *status.State
		}
	}
	statusesJson, err := json.Marshal(allStatusesMap)
	fields := map[string]string{
		"STATUSES":      allStatuses,
		"STATUSES_JSON": string(statusesJson),
	}
	return fields
}
