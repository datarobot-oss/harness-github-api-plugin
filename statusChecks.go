package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v62/github"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
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

func waitForStatus(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string,
	sha string, statusName string, timeoutMinutes string) map[string]string {
	var result string
	timeoutMinutesInt, errConv := strconv.Atoi(timeoutMinutes)
	failOnErr(errConv)
	// 1 attempt is 30 seconds
	maxAttempts := timeoutMinutesInt * 2
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		statuses, _, err := client.Repositories.ListStatuses(*ctx, repositoryOwner, repositoryName, sha, nil)
		failOnErr(err)
		slices.Reverse(statuses)
		for _, status := range statuses {
			if *status.Context == statusName {
				result = *status.State
			}
		}
		if result == "" {
			fmt.Println("No status check with name " + statusName)
		} else {
			fmt.Println(statusName + " " + result)
		}
		if result == "success" || result == "failure" {
			break
		}
		time.Sleep(30 * time.Second)
	}

	if result == "pending" {
		log.Fatal("max attempts reached but result is still pending")
	} else if result == "" {
		log.Fatal("max attempts reached but status check never appears")
	}
	return map[string]string{"RESULT": result}
}

func listStatusChecks(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, sha string) map[string]string {
	statuses, _, err := client.Repositories.ListStatuses(*ctx, repositoryOwner, repositoryName, sha, nil)
	failOnErr(err)
	allStatuses := ""
	var allStatusesArray []map[string]string
	for i, status := range statuses {
		prefix := ""
		if i > 0 {
			prefix = ", "
		}
		if !strings.Contains(allStatuses, *status.Context) {
			allStatuses += prefix + *status.Context + " " + *status.State
			allStatusesArray = append(allStatusesArray, map[string]string{
				"context": *status.Context,
				"status":  *status.State,
			})
		}
	}
	statusesJson, err := json.Marshal(allStatusesArray)
	fields := map[string]string{
		"STATUSES":      allStatuses,
		"STATUSES_JSON": string(statusesJson),
	}
	return fields
}
