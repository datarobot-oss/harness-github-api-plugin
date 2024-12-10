# harness-github-api-plugin

## Description

This plugin is designed to be the Swiss Army knife for git API operations in Harness pipelines.

## Currently supported features

 - Get pull request details
 - Get changed files
 - Create a pull request
 - Merge a pull request
 - Create/update git tag
 - Set GitHub status check
 - TBD, if you need a new feature added, reach to us in slack #continuous-integration

## How the plugin works

Plugin takes a list of settings (environment variables):
 - common - required for all operations
 - variable COMMANDS with list of actions to perform
 - command-specific variables

It iterates through commands, performs requested operations and optionally exports data to the 'step output' harness section.

## Common plugin settings
`GITHUB_AUTH_TOKEN` - The GH PAT used for API calls \
`REPOSITORY_NAME` \
`REPOSITORY_OWNER` (GitHub organization, e.g. datarobot for repository) \
`COMMANDS` - The list of actions the plugin should perform

## Commands and commands-specific settings
### Get pull request details
`COMMANDS=”getPrDetails”` \
`PR_NUMBER=123`

### Get changed files
`COMMANDS=”getChangedFiles”` \
`PR_NUMBER=123`


### Create pull request
`COMMANDS=“createPullRequest”` \
`PR_SOURCE_BRANCH=author/ticket/feature` \
`PR_TARGET_BRANCH=main` \
`PR_TITLE=”[CIIT-321] Update version file”` \
`PR_BODY=”Update version to v0.0.3”` 

### Merge pull request
`COMMANDS=”mergePr”` \
`PR_NUMBER=123` \
`MERGE_COMMENT="Updating artifact version to 0.1.321"`

### Set tag
`COMMANDS=“setTag”` \
`TAG_NAME=latest_green_staging` \
`SHA=<+{path to variable with git sha}>` 

### Set status check
`COMMANDS="setStatusCheck"` \
`STATUS_CHECK_SHA=<+{path to variable with git sha}>` \
`STATUS_CHECK_CONTEXT=BuildAndPushImages` \
`STATUS_CHECK_STATUS=success` \
`STATUS_CHECK_URL=<+{pipeline/stage URL}>` \
`STATUS_CHECK_DESCRIPTION="Build and push project images"` 

### Get status checks
`COMMANDS="getStatuses"` \
`REF=branch-name-or-sha`

## Caveats and restrictions
Harness has a number of limitations that are related to plugin functionality:
- It doesn't process multiline output data
- It limits length of output data

If you hit these limitations, you can read the file named `outputVariables` that is created by the plugin in the current workspace directory. 
Within the same pipeline stage run the command `source outputVariables`.



## Links

Plugin docker image: https://ghcr.io/datarobot-oss/harness-github-api-plugin \
Harness documentation for plugins: https://developer.harness.io/docs/category/use-plugins/ 
