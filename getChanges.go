package main

import (
	"context"
	"encoding/json"
	"github.com/google/go-github/v62/github"
	"github.com/waigani/diffparser"
	"regexp"
	"strconv"
)

func getChanges(client *github.Client, ctx *context.Context, repositoryName string, repositoryOwner string, base string, head string) map[string]string {
	comp, _, _ := client.Repositories.CompareCommitsRaw(*ctx, repositoryOwner, repositoryName, base, head, github.RawOptions{Type: github.Diff})
	diff, _ := diffparser.Parse(comp)

	fields := map[string]string{
		"CHANGED_FILES":        "",
		"MODIFIED_FILES":       "",
		"ADDED_FILES":          "",
		"REMOVED_FILES":        "",
		"CHANGED_FILES_NUMBER": "",
	}

	removedFiles := []string{}
	modifiedFiles := []string{}
	addedFiles := []string{}
	changedFiles := []string{}

	for _, file := range diff.Files {
		fileName := file.NewName
		if fileName == "" {
			fileName = file.OrigName
		}
		if fileName == "" {
			re := regexp.MustCompile("^.*\\sb/")
			fileName = re.ReplaceAllString(file.DiffHeader, "")
		}
		changedFiles = append(changedFiles, fileName)
		switch fileStatus := file.Mode; fileStatus {
		case 0:
			removedFiles = append(removedFiles, fileName)
		case 1:
			modifiedFiles = append(modifiedFiles, fileName)
		case 2:
			addedFiles = append(addedFiles, fileName)
		}

		removedFilesBytes, _ := json.Marshal(removedFiles)
		fields["REMOVED_FILES"] = string(removedFilesBytes)

		modifiedFilesBytes, _ := json.Marshal(modifiedFiles)
		fields["MODIFIED_FILES"] = string(modifiedFilesBytes)

		addedFilesBytes, _ := json.Marshal(addedFiles)
		fields["ADDED_FILES"] = string(addedFilesBytes)

		changedFilesBytes, _ := json.Marshal(changedFiles)
		fields["CHANGED_FILES"] = string(changedFilesBytes)

		fields["CHANGED_FILES_NUMBER"] = strconv.Itoa(len(diff.Files))
	}

	return fields
}
