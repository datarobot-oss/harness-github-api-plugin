package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v62/github"
)

func updateCreateTag(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, tagName string, sha string) {
	_, _, err := client.Git.GetRef(*ctx, repositoryOwner, repositoryName, tagName)
	if err == nil {
		fmt.Println("Updating existing tag " + tagName + " " + sha)
		_, _, err = client.Git.UpdateRef(*ctx, repositoryOwner, repositoryName, &github.Reference{Ref: &tagName, Object: &github.GitObject{SHA: &sha}}, true)
	} else {
		fmt.Println("Creating new tag" + tagName + " " + sha)
		_, _, err = client.Git.CreateRef(*ctx, repositoryOwner, repositoryName, &github.Reference{Ref: &tagName, Object: &github.GitObject{SHA: &sha}})
	}
	failOnErr(err)
}
