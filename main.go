package main

import (
	"context"
	"fmt"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v62/github"
	"os"
	"os/exec"
	"strings"
)

var (
	token           = os.Getenv("PLUGIN_GITHUB_AUTH_TOKEN")
	repositoryName  = os.Getenv("PLUGIN_REPOSITORY_NAME")
	repositoryOwner = os.Getenv("PLUGIN_REPOSITORY_OWNER")
	commands        = os.Getenv("PLUGIN_COMMANDS")
	outputFile      = os.Getenv("DRONE_OUTPUT")
	copyOutputFile  = "outputVariables" // TODO: make the path configurable
)

func main() {
	verifyPluginParameters([]string{"PLUGIN_GITHUB_AUTH_TOKEN", "PLUGIN_REPOSITORY_NAME", "PLUGIN_REPOSITORY_OWNER", "PLUGIN_COMMANDS"})
	ctx := context.Background()
	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(nil)
	failOnErr(err)
	client := github.NewClient(rateLimiter).WithAuthToken(token)

	results, err := os.Create(outputFile)
	failOnErr(err)
	if strings.Contains(commands, "getPrDetails") {
		verifyPluginParameters([]string{"PLUGIN_PR_NUMBER"})
		fields := getPullRequest(client, &ctx, repositoryName, repositoryOwner, os.Getenv("PLUGIN_PR_NUMBER"))
		writeResult(*results, fields)
	}
	if strings.Contains(commands, "getChangedFiles") {
		verifyPluginParameters([]string{"PLUGIN_PR_NUMBER"})
		prFields := getPullRequest(client, &ctx, repositoryName, repositoryOwner, os.Getenv("PLUGIN_PR_NUMBER"))
		//writeResult(*results, prFields)
		fields := getChanges(client, &ctx, repositoryName, repositoryOwner, prFields["BASE_SHA"], prFields["HEAD_SHA"])
		writeResult(*results, fields)
	}
	if strings.Contains(commands, "setTag") {
		verifyPluginParameters([]string{"PLUGIN_TAG_NAME", "PLUGIN_SHA"})
		tagName := "refs/tags/" + os.Getenv("PLUGIN_TAG_NAME")
		updateCreateTag(client, &ctx, repositoryName, repositoryOwner, tagName, os.Getenv("PLUGIN_SHA"))
	}
	if strings.Contains(commands, "createPullRequest") {
		verifyPluginParameters([]string{"PLUGIN_PR_SOURCE_BRANCH", "PLUGIN_PR_TARGET_BRANCH", "PLUGIN_PR_TITLE", "PLUGIN_PR_BODY"})
		fields := createPullRequest(client, &ctx, repositoryName, repositoryOwner,
			os.Getenv("PLUGIN_PR_SOURCE_BRANCH"),
			os.Getenv("PLUGIN_PR_TARGET_BRANCH"),
			os.Getenv("PLUGIN_PR_TITLE"),
			os.Getenv("PLUGIN_PR_BODY"))
		writeResult(*results, fields)
	}
	if strings.Contains(commands, "setStatusCheck") {
		verifyPluginParameters([]string{"PLUGIN_STATUS_CHECK_SHA",
			"PLUGIN_STATUS_CHECK_CONTEXT",
			"PLUGIN_STATUS_CHECK_STATUS",
			"PLUGIN_STATUS_CHECK_URL",
			"PLUGIN_STATUS_CHECK_DESCRIPTION"})
		setStatusCheck(client, &ctx, repositoryName, repositoryOwner,
			os.Getenv("PLUGIN_STATUS_CHECK_SHA"),
			os.Getenv("PLUGIN_STATUS_CHECK_CONTEXT"),
			os.Getenv("PLUGIN_STATUS_CHECK_STATUS"),
			os.Getenv("PLUGIN_STATUS_CHECK_URL"),
			os.Getenv("PLUGIN_STATUS_CHECK_DESCRIPTION"))
	}
	if strings.Contains(commands, "getStatuses") {
		verifyPluginParameters([]string{"PLUGIN_REF"})
		fields := listStatusChecks(client, &ctx, repositoryName, repositoryOwner, os.Getenv("PLUGIN_REF"))
		writeResult(*results, fields)
	}
	if strings.Contains(commands, "waitForStatus") {
		verifyPluginParameters([]string{"PLUGIN_REF", "PLUGIN_STATUS_CONTEXT"})
		fields := waitForStatus(client, &ctx, repositoryName, repositoryOwner,
			os.Getenv("PLUGIN_REF"),
			os.Getenv("PLUGIN_STATUS_CHECK_CONTEXT"),
			os.Getenv("PLUGIN_STATUS_CHECK_WAIT_TIMEOUT"))
		writeResult(*results, fields)
	}
	if strings.Contains(commands, "mergePr") {
		verifyPluginParameters([]string{"PLUGIN_PR_NUMBER"})
		mergePullRequest(client, &ctx, repositoryName, repositoryOwner,
			os.Getenv("PLUGIN_PR_NUMBER"),
			os.Getenv("PLUGIN_MERGE_COMMENT"))
	}
	results.Close()

	out, err := exec.Command("cp", "-v", outputFile, copyOutputFile).Output()
	failOnErr(err)
	fmt.Print(string(out[:]))
}
