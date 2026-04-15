package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/autelrobotics/codearts-cli/internal/client"
	"github.com/autelrobotics/codearts-cli/internal/core"
	"github.com/autelrobotics/codearts-cli/internal/output"
)

func newRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "CodeArts Repo (code hosting) operations",
	}
	cmd.AddCommand(newRepoMRCmd())
	return cmd
}

func newRepoMRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mr",
		Short: "Merge request operations",
	}
	cmd.AddCommand(newRepoMRCommentCmd())
	return cmd
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
			if _, err := fmt.Sscanf(args[0], "%d", &o.repoID); err != nil || o.repoID <= 0 {
				return fmt.Errorf("repository_id must be a positive integer, got %q", args[0])
			}
			if _, err := fmt.Sscanf(args[1], "%d", &o.mrIID); err != nil || o.mrIID <= 0 {
				return fmt.Errorf("merge_request_iid must be a positive integer, got %q", args[1])
			}
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
	rawBody, err := firstNonEmpty("--body-json", o.bodyJSON, "--body-file", o.bodyFile)
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
