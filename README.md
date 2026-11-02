# harness-github-api-plugin

## Description

This plugin is designed to be the Swiss Army knife for git API operations in Harness pipelines.

## Currently supported features

 - Get pull request details
 - Get changed files
 - Create a pull request
 - Add labels to a pull request
 - Merge a pull request
 - Create or update a pull request comment
 - Add a reaction to a comment
 - Get a user's permission on the repository
 - Create/update git tag
 - Get a git reference
 - Set GitHub status check
 - Get status checks
 - Wait for a status check
 - TBD, if you need a new feature added, reach to us in slack #continuous-integration

## How the plugin works

Plugin takes a list of settings (environment variables):
 - common - required for all operations
 - variable COMMANDS with list of actions to perform
 - command-specific variables

It iterates through commands, performs requested operations and optionally exports data to the 'step output' harness section.

Several commands can be listed in `COMMANDS` (for example `COMMANDS="getPrDetails,getChangedFiles"`).
They always run in the order the sections below are listed in, not in the order they are written in `COMMANDS`.
If you need a specific order, use separate steps.

## Common plugin settings
`GITHUB_AUTH_TOKEN` - The GH PAT used for API calls \
`REPOSITORY_NAME` \
`REPOSITORY_OWNER` (GitHub organization, e.g. datarobot for repository) \
`COMMANDS` - The list of actions the plugin should perform

## Example Harness step

```yaml
- step:
    type: Plugin
    name: CommentOnPr
    identifier: CommentOnPr
    spec:
      connectorRef: <docker registry connector for ghcr.io>
      image: ghcr.io/datarobot-oss/harness-github-api-plugin
      settings:
        GITHUB_AUTH_TOKEN: <+secrets.getValue("github_token")>
        REPOSITORY_OWNER: my-org
        REPOSITORY_NAME: my-repository
        COMMANDS: upsertPrComment
        PR_NUMBER: <+trigger.prNumber>
        COMMENT_MARKER: deploy-status
        COMMENT_BODY: |-
          :rocket: Deployment started: [execution](<+pipeline.executionUrl>)
```

Outputs are read from later steps as `<+execution.steps.CommentOnPr.output.outputVariables.COMMENT_ID>`.

## Commands and commands-specific settings
### Get pull request details
`COMMANDS="getPrDetails"` \
`PR_NUMBER=123`

Outputs: `TITLE`, `BODY`, `USER_LOGIN`, `USER_EMAIL`, `BASE_BRANCH_NAME`, `BASE_SHA`, `HEAD_BRANCH_NAME`, `HEAD_SHA`,
`MERGE_COMMIT_SHA`, `STATE`, `URL` (API URL of the pull request), `CREATED_AT`, `IS_DRAFT`

### Get changed files
`COMMANDS="getChangedFiles"` \
`PR_NUMBER=123`

Outputs (JSON arrays of file paths): `CHANGED_FILES`, `MODIFIED_FILES`, `ADDED_FILES`, `REMOVED_FILES`, and `CHANGED_FILES_NUMBER`

### Set tag
`COMMANDS="setTag"` \
`TAG_NAME=latest_green_staging` \
`SHA=<+{path to variable with git sha}>`

Creates the tag, or moves it to `SHA` if it already exists.

### Create pull request
`COMMANDS="createPullRequest"` \
`PR_SOURCE_BRANCH=author/ticket/feature` \
`PR_TARGET_BRANCH=main` \
`PR_TITLE="[CIIT-321] Update version file"` \
`PR_BODY="Update version to v0.0.3"` \
`PR_LABELS="label1, label2"` (optional)

If a pull request for the same branches already exists, its title and body are updated instead.

Outputs: `PR_NUMBER`, `PR_URL`, `PR_TITLE`, `PR_HEAD`, `PR_HEAD_SHA`, `PR_BASE`

### Add pull request labels
`COMMANDS="AddPullRequestLabels"` \
`PR_NUMBER=123` \
`PR_LABELS="label1, label2"`

### Set status check
`COMMANDS="setStatusCheck"` \
`STATUS_CHECK_SHA=<+{path to variable with git sha}>` \
`STATUS_CHECK_CONTEXT=BuildAndPushImages` \
`STATUS_CHECK_STATUS=success` (one of `pending`, `success`, `failure`, `error`) \
`STATUS_CHECK_URL=<+{pipeline/stage URL}>` \
`STATUS_CHECK_DESCRIPTION="Build and push project images"`

### Get status checks
`COMMANDS="getStatuses"` \
`REF=branch-name-or-sha`

Outputs: `STATUSES` (`"<context> <state>, ..."`), `STATUSES_JSON` (`[{"context": ..., "status": ...}]`), latest state per context

### Wait for status check
`COMMANDS="waitForStatus"` \
`REF=branch-name-or-sha` \
`STATUS_CHECK_CONTEXT=BuildAndPushImages` \
`STATUS_CHECK_WAIT_TIMEOUT=30` (minutes, required)

Polls every 30 seconds until the status is `success` or `failure`, and fails the step if it is still pending
or missing when the timeout runs out.

Outputs: `RESULT`

### Merge pull request
`COMMANDS="mergePr"` \
`PR_NUMBER=123` \
`MERGE_COMMENT="Updating artifact version to 0.1.321"`

### Get git reference
`COMMANDS="getRef"` \
`REF=heads/main` (or `tags/v1.0.0`)

Outputs: `REF`, `SHA`, `SHORT_SHA` (11 characters), `URL`

### Get user permission
`COMMANDS="getUserPermission"` \
`USERNAME=octocat`

Reports the user's permission on the repository, for example to check that the author of a PR comment
may start a pipeline. A user GitHub does not know is reported as `none`; the step does not fail.

Outputs: `PERMISSION` (`admin`, `write`, `read` or `none`), `HAS_WRITE_ACCESS` (`true` for `admin` and `write`)

### Create or update pull request comment
`COMMANDS="upsertPrComment"` \
`PR_NUMBER=123` \
`COMMENT_BODY="Deployment finished"` \
`COMMENT_BODY_FILE=report.md` (optional, used instead of `COMMENT_BODY`) \
`COMMENT_MARKER=deploy-status` (optional)

Without `COMMENT_MARKER`, every run adds a new comment.

With `COMMENT_MARKER`, the plugin adds a hidden marker (`<!-- harness-github-api-plugin:<marker> -->`) to the body.
On later runs it edits the most recent comment carrying the same marker instead of adding another one, so a pipeline
can keep one comment per pull request up to date. Only comments written by the owner of `GITHUB_AUTH_TOKEN` are
considered. Use different markers for comments that should stay separate, for example `deploy-eks` and `deploy-gke`.

`COMMENT_BODY_FILE` is read from the step's workspace, so an earlier `Run` step in the same stage can write
a long or multiline report to it.

Optional settings set to `null` (what Harness passes for an expression without a value) are treated as unset.

Outputs: `COMMENT_ID`, `COMMENT_URL`, `COMMENT_ACTION` (`created` or `updated`)

### Add reaction to a comment
`COMMANDS="addReaction"` \
`COMMENT_ID=1234567890` \
`REACTION=eyes` (one of `+1`, `-1`, `laugh`, `confused`, `heart`, `hooray`, `rocket`, `eyes`)

Works for pull request and issue comments, for example to acknowledge the comment that triggered a pipeline
(`<+trigger.payload.comment.id>`). Adding a reaction that is already there is not an error.

## Caveats and restrictions
Harness has a number of limitations that are related to plugin functionality:
- It doesn't process multiline output data
- It limits length of output data

If you hit these limitations, you can read the file named `outputVariables` that is created by the plugin in the current workspace directory.
Within the same pipeline stage run the command `source outputVariables`.

Any failed API call fails the step.

## Links

Plugin docker image: https://ghcr.io/datarobot-oss/harness-github-api-plugin \
Harness documentation for plugins: https://developer.harness.io/docs/category/use-plugins/
