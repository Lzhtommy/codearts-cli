package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Lzhtommy/codearts-cli/internal/client"
	"github.com/Lzhtommy/codearts-cli/internal/core"
	"github.com/Lzhtommy/codearts-cli/internal/output"
)

// ParseRepoID rejects anything that isn't a pure positive integer (no
// hex, no leading signs, no whitespace). Earlier this used fmt.Sscanf
// which silently truncated UUIDs to their leading digit run — e.g. the
// UUID "759278ab..." quietly parsed as 759278. That would quietly send
// the request to the wrong repository, so we reject non-numeric input
// up-front.
func ParseRepoID(raw string) (int, error) {
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf(
			"repository_id must be a positive integer, got %q — note that repo_id is the numeric repo ID (visible in CodeArts Repo console), NOT the 32-char project UUID",
			raw,
		)
	}
	return v, nil
}

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "CodeArts Repo (code hosting) operations",
	}
	cmd.AddCommand(newRepoListCmd())
	cmd.AddCommand(newRepoMRCmd())
	cmd.AddCommand(newRepoMemberCmd())
	return cmd
}

// ----------------------- repo member -----------------------

func newRepoMemberCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Repository member operations",
	}
	cmd.AddCommand(newRepoMemberListCmd())
	return cmd
}

type memberListOpts struct {
	repoID     int
	search     string
	offset     int
	limit      int
	permission string
	action     string
	dryRun     bool
}

func newRepoMemberListCmd() *cobra.Command {
	o := &memberListOpts{}
	cmd := &cobra.Command{
		Use:   "list <repository_id>",
		Short: "List members of a repository (ListMembers API)",
		Long: `List members of a CodeArts repository.

<repository_id> is the numeric repository ID (int), not a UUID. Get it
from ` + "`codearts-cli repo list --project-id <proj>`" + `.

Returned fields include user_id, user_name, user_nick_name, tenant_name,
repository_role_name, and service_license_status (0=stopped, 1=active).

Filter by permission point + action (both enum):
    permission: repository | code | member | branch | tag | mr | label
    action    : per-permission, e.g. code→push/download,
                                  mr→create/update/comment/review/approve/merge/close/reopen,
                                  repository→create/fork/delete/setting,
                                  member/branch/tag/label→create/update/delete

EXAMPLES:
    # List all members
    codearts-cli repo member list 8147520

    # Search by name / nickname / tenant
    codearts-cli repo member list 8147520 --search "zhang"

    # Paginate
    codearts-cli repo member list 8147520 --offset 20 --limit 50

    # Only members with push permission on code
    codearts-cli repo member list 8147520 --permission code --action push

API reference: https://support.huaweicloud.com/api-codeartsrepo/ListMembers.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := ParseRepoID(args[0])
			if err != nil {
				return err
			}
			o.repoID = v
			return runRepoMemberList(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.search, "search", "", "Fuzzy match on user_name / nick_name / tenant_name")
	cmd.Flags().IntVar(&o.offset, "offset", 0, "Pagination offset (0-based, 0 = API default)")
	cmd.Flags().IntVar(&o.limit, "limit", 0, "Results per page (1-100, 0 = API default 20)")
	cmd.Flags().StringVar(&o.permission, "permission", "", "Permission point filter (repository|code|member|branch|tag|mr|label)")
	cmd.Flags().StringVar(&o.action, "action", "", "Permission action filter (depends on --permission)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runRepoMemberList(cmd *cobra.Command, o *memberListOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if (o.action != "") && (o.permission == "") {
		return fmt.Errorf("--action requires --permission (the action enum is scoped to a permission point)")
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		q := map[string]interface{}{}
		if o.search != "" {
			q["search"] = o.search
		}
		if o.offset > 0 {
			q["offset"] = o.offset
		}
		if o.limit > 0 {
			q["limit"] = o.limit
		}
		if o.permission != "" {
			q["permission"] = o.permission
		}
		if o.action != "" {
			q["action"] = o.action
		}
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":        "GET",
			"path":          fmt.Sprintf("/v4/repositories/%d/members", o.repoID),
			"repository_id": o.repoID,
			"query":         q,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListMembers(context.Background(), o.repoID, &client.ListMembersRequest{
		Search:     o.search,
		Offset:     o.offset,
		Limit:      o.limit,
		Permission: o.permission,
		Action:     o.action,
	})
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- repo list -----------------------

type repoListOpts struct {
	projectID string
	pageIndex int
	pageSize  int
	search    string
	dryRun    bool
}

func newRepoListCmd() *cobra.Command {
	o := &repoListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repositories in a project (ShowAllRepositoryByTwoProjectId)",
		Long: `List repositories in a CodeArts project.

--project-id is required (32-char project UUID).

The response includes each repo's numeric repository_id — use that for
repo mr create / repo mr comment commands.

EXAMPLES:
    # List all repos
    codearts-cli repo list --project-id <proj>

    # Search by name
    codearts-cli repo list --project-id <proj> --search "backend"

    # Paginate
    codearts-cli repo list --project-id <proj> --page-index 2 --page-size 10

API reference: https://support.huaweicloud.com/api-codeartsrepo/ShowAllRepositoryByTwoProjectId.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepoList(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "(required) CodeArts project UUID (extract from git remote URL)")
	cmd.Flags().IntVar(&o.pageIndex, "page-index", 0, "Page number (1-based, 0 = API default)")
	cmd.Flags().IntVar(&o.pageSize, "page-size", 0, "Results per page (1-100, 0 = API default 20)")
	cmd.Flags().StringVar(&o.search, "search", "", "Search by repo name or creator name")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runRepoList(cmd *cobra.Command, o *repoListOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := o.projectID
	if projectID == "" {
		return fmt.Errorf("--project-id is required for repo commands")
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		q := map[string]interface{}{}
		if o.pageIndex > 0 {
			q["page_index"] = o.pageIndex
		}
		if o.pageSize > 0 {
			q["page_size"] = o.pageSize
		}
		if o.search != "" {
			q["search"] = o.search
		}
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "GET",
			"path":       fmt.Sprintf("/v2/projects/%s/repositories", projectID),
			"project_id": projectID,
			"query":      q,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListRepositories(context.Background(), projectID, o.pageIndex, o.pageSize, o.search)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

func newRepoMRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mr",
		Short: "Merge request operations",
	}
	cmd.AddCommand(newRepoMRCreateCmd())
	cmd.AddCommand(newRepoMRCommentCmd())
	return cmd
}

// ----------------------- repo mr create -----------------------

type mrCreateOpts struct {
	repoID               int
	title                string
	sourceBranch         string
	targetBranch         string
	description          string
	targetRepoID         int
	reviewerIDs          string
	assigneeIDs          string
	approvalReviewerIDs  string
	approvalApproversIDs string
	milestoneID          int
	forceRemoveSource    bool
	squash               bool
	squashCommitMessage  string
	workItemIDs          []string
	onlyAssigneeMerge    bool
	bodyJSON             string
	bodyFile             string
	dryRun               bool
}

func newRepoMRCreateCmd() *cobra.Command {
	o := &mrCreateOpts{}
	cmd := &cobra.Command{
		Use:   "create <repository_id>",
		Short: "Create a merge request (CreateMergeRequest API)",
		Long: `Create a merge request on a repository.

<repository_id> is the numeric repository ID (int), not a UUID.

Minimum:
    codearts-cli repo mr create 12345 \
      --title "feat: my change" \
      --source feat/x --target main

Advanced (reviewers / approvers / squash / work items):
    codearts-cli repo mr create 12345 \
      --title "..." --source feat/x --target main \
      --reviewers "user_id_a,user_id_b" \
      --assignees "user_id_c" \
      --squash --force-remove-source \
      --work-item 1251275102548402177

Or pass the whole JSON body:
    codearts-cli repo mr create 12345 --body-file mr.json

API reference: https://support.huaweicloud.com/api-codeartsrepo/CreateMergeRequest.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := ParseRepoID(args[0])
			if err != nil {
				return err
			}
			o.repoID = v
			return runMRCreate(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.title, "title", "", "MR title (required)")
	cmd.Flags().StringVar(&o.sourceBranch, "source", "", "source_branch (required)")
	cmd.Flags().StringVar(&o.targetBranch, "target", "", "target_branch (required)")
	cmd.Flags().StringVar(&o.description, "description", "", "MR description")
	cmd.Flags().IntVar(&o.targetRepoID, "target-repo-id", 0, "target_repository_id (cross-repo MR)")
	cmd.Flags().StringVar(&o.reviewerIDs, "reviewers", "", "Comma-separated reviewer user_ids")
	cmd.Flags().StringVar(&o.assigneeIDs, "assignees", "", "Comma-separated assignee user_ids")
	cmd.Flags().StringVar(&o.approvalReviewerIDs, "approval-reviewers", "", "Comma-separated approval reviewer user_ids")
	cmd.Flags().StringVar(&o.approvalApproversIDs, "approval-approvers", "", "Comma-separated approver user_ids")
	cmd.Flags().IntVar(&o.milestoneID, "milestone-id", 0, "milestone_id")
	cmd.Flags().BoolVar(&o.forceRemoveSource, "force-remove-source", false, "Auto-delete source branch on merge")
	cmd.Flags().BoolVar(&o.squash, "squash", false, "Squash commits on merge")
	cmd.Flags().StringVar(&o.squashCommitMessage, "squash-message", "", "Squash commit message (only with --squash)")
	cmd.Flags().StringSliceVar(&o.workItemIDs, "work-item", nil, "Associated work item id (repeatable; comma-separated also works)")
	cmd.Flags().BoolVar(&o.onlyAssigneeMerge, "only-assignee-merge", false, "Restrict merge to assignees only")
	cmd.Flags().StringVar(&o.bodyJSON, "body-json", "", "Full JSON body (overrides flag-based fields)")
	cmd.Flags().StringVar(&o.bodyFile, "body-file", "", "Path to a JSON file for the full body")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runMRCreate(cmd *cobra.Command, o *mrCreateOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	var body interface{}
	rawBody, err := FirstNonEmpty("--body-json", o.bodyJSON, "--body-file", o.bodyFile)
	if err != nil {
		return err
	}
	if rawBody != "" {
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(rawBody), &m); err != nil {
			return fmt.Errorf("parse body JSON: %w", err)
		}
		body = m
	} else {
		if o.title == "" || o.sourceBranch == "" || o.targetBranch == "" {
			return fmt.Errorf("--title, --source and --target are all required (or pass --body-json / --body-file)")
		}
		// Flatten repeatable / comma-separated --work-item into a single
		// flat slice, matching the behavior of `issue batch-update --id`.
		var flatItems []string
		for _, entry := range o.workItemIDs {
			for _, s := range strings.Split(entry, ",") {
				if v := strings.TrimSpace(s); v != "" {
					flatItems = append(flatItems, v)
				}
			}
		}
		body = &client.CreateMRRequest{
			Title:                   o.title,
			SourceBranch:            o.sourceBranch,
			TargetBranch:            o.targetBranch,
			Description:             o.description,
			TargetRepositoryID:      o.targetRepoID,
			ReviewerIDs:             o.reviewerIDs,
			AssigneeIDs:             o.assigneeIDs,
			ApprovalReviewerIDs:     o.approvalReviewerIDs,
			ApprovalApproversIDs:    o.approvalApproversIDs,
			MilestoneID:             o.milestoneID,
			ForceRemoveSourceBranch: o.forceRemoveSource,
			Squash:                  o.squash,
			SquashCommitMessage:     o.squashCommitMessage,
			WorkItemIDs:             flatItems,
			OnlyAssigneeCanMerge:    o.onlyAssigneeMerge,
		}
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":        "POST",
			"path":          fmt.Sprintf("/v4/repositories/%d/merge-requests", o.repoID),
			"repository_id": o.repoID,
			"body":          body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.CreateMergeRequest(context.Background(), o.repoID, body)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Merge request created on repo %d", o.repoID)
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- repo mr comment -----------------------

type mrCommentOpts struct {
	repoID       int
	mrIID        int
	body         string
	severity     string
	assigneeID   string
	categories   string
	modules      string
	proposerID   string
	lineTypes    string
	bodyFile     string
	bodyJSON     string
	dryRun       bool
}

func newRepoMRCommentCmd() *cobra.Command {
	o := &mrCommentOpts{}
	cmd := &cobra.Command{
		Use:   "comment <repository_id> <merge_request_iid>",
		Short: "Post a review discussion on a merge request (CreateMergeRequestDiscussion)",
		Long: `Create a merge-request discussion (i.e. a code review comment).

<repository_id> and <merge_request_iid> are both integers (numeric IDs).

Simplest usage:
    codearts-cli repo mr comment 12345 7 --body "Please add a unit test here"

For line-level comments or richer review metadata (position, review_categories,
etc.), pass --body-file with the full JSON.

API reference: https://support.huaweicloud.com/api-codeartsrepo/CreateMergeRequestDiscussion.html`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := ParseRepoID(args[0])
			if err != nil {
				return err
			}
			o.repoID = v
			iid, err := strconv.Atoi(args[1])
			if err != nil || iid <= 0 {
				return fmt.Errorf("merge_request_iid must be a positive integer, got %q", args[1])
			}
			o.mrIID = iid
			return runMRComment(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.body, "body", "", "Comment content")
	cmd.Flags().StringVar(&o.severity, "severity", "", "suggestion | minor | major | fatal")
	cmd.Flags().StringVar(&o.assigneeID, "assignee-id", "", "Assignee user ID (optional)")
	cmd.Flags().StringVar(&o.categories, "review-categories", "", "Review categories (optional)")
	cmd.Flags().StringVar(&o.modules, "review-modules", "", "Review modules (optional)")
	cmd.Flags().StringVar(&o.proposerID, "proposer-id", "", "Reviewer ID (optional)")
	cmd.Flags().StringVar(&o.lineTypes, "line-types", "", "Line type (optional; for line-level comments use --body-file)")
	cmd.Flags().StringVar(&o.bodyJSON, "body-json", "", "Full JSON body (overrides flag-based fields)")
	cmd.Flags().StringVar(&o.bodyFile, "body-file", "", "Path to a JSON file for the full body")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runMRComment(cmd *cobra.Command, o *mrCommentOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	var body interface{}
	rawBody, err := FirstNonEmpty("--body-json", o.bodyJSON, "--body-file", o.bodyFile)
	if err != nil {
		return err
	}
	if rawBody != "" {
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(rawBody), &m); err != nil {
			return fmt.Errorf("parse body JSON: %w", err)
		}
		body = m
	} else {
		if o.body == "" {
			return fmt.Errorf("--body is required (or pass --body-json / --body-file)")
		}
		body = &client.CreateMRDiscussionRequest{
			Body:             o.body,
			Severity:         o.severity,
			AssigneeID:       o.assigneeID,
			ReviewCategories: o.categories,
			ReviewModules:    o.modules,
			ProposerID:       o.proposerID,
			LineTypes:        o.lineTypes,
		}
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":            "POST",
			"path":              fmt.Sprintf("/v4/repositories/%d/merge-requests/%d/discussions", o.repoID, o.mrIID),
			"repository_id":     o.repoID,
			"merge_request_iid": o.mrIID,
			"body":              body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.CreateMergeRequestDiscussion(context.Background(), o.repoID, o.mrIID, body)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Discussion posted on MR !%d (repo %d)", o.mrIID, o.repoID)
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}
