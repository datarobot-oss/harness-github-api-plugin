package main

import (
	"context"
	"github.com/google/go-github/v62/github"
)

func getRef(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, ref string) map[string]string {
	refResponse, _, err := client.Git.GetRef(*ctx, repositoryOwner, repositoryName, ref)
	failOnErr(err)
	fields := map[string]string{
		"REF":       refResponse.GetRef(),
		"SHA":       refResponse.Object.GetSHA(),
		"SHORT_SHA": refResponse.Object.GetSHA()[0:11],
		"URL":       refResponse.GetURL(),
	}
	return fields
}
