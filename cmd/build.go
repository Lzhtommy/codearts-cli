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

func newBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "CodeArts Build (编译构建 / CodeCI) operations",
	}
	cmd.AddCommand(newBuildListCmd())
	cmd.AddCommand(newBuildRunCmd())
	cmd.AddCommand(newBuildStopCmd())
	return cmd
}

// ------------------------------ build list ------------------------------

type buildListOpts struct {
	projectID   string
	pageIndex   int
	pageSize    int
	search      string
	sortField   string
	sortOrder   string
	creatorID   string
	buildStatus string
	byGroup     bool
	groupPathID string
	dryRun      bool
}

func newBuildListCmd() *cobra.Command {
	o := &buildListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List build jobs in a project (ListProjectJobs API)",
		Long: `List CodeArts Build jobs in a project.

--project-id is required (32-char project UUID). The returned job ` + "`id`" + `
is the identifier used by ` + "`build run`" + ` / ` + "`build stop`" + `.

EXAMPLES:
    # List all jobs
    codearts-cli build list --project-id <proj>

    # Search by name + paginate
    codearts-cli build list --project-id <proj> --search "backend" \
      --page-index 0 --page-size 50

    # Filter by last-build status
    codearts-cli build list --project-id <proj> --build-status red

API reference: https://support.huaweicloud.com/api-codeci/ListProjectJobs.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuildList(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "(required) CodeArts project UUID")
	cmd.Flags().IntVar(&o.pageIndex, "page-index", 0, "Pagination starting page (0-based, 0 = API default)")
	cmd.Flags().IntVar(&o.pageSize, "page-size", 0, "Items per page (1-100, 0 = API default 10)")
	cmd.Flags().StringVar(&o.search, "search", "", "Fuzzy query on job name / creator")
	cmd.Flags().StringVar(&o.sortField, "sort-field", "", "Sort field")
	cmd.Flags().StringVar(&o.sortOrder, "sort-order", "", "Sort direction (asc | desc)")
	cmd.Flags().StringVar(&o.creatorID, "creator-id", "", "Filter by creator user_id")
	cmd.Flags().StringVar(&o.buildStatus, "build-status", "", "Filter by last-build status: red | blue | timeout | aborted | building | none")
	cmd.Flags().BoolVar(&o.byGroup, "by-group", false, "Enable grouping (pairs with --group-path-id)")
	cmd.Flags().StringVar(&o.groupPathID, "group-path-id", "", "Group path ID (used when --by-group is set)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runBuildList(cmd *cobra.Command, o *buildListOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if o.projectID == "" {
		return fmt.Errorf("--project-id is required for build commands")
	}

	req := &client.ListProjectJobsRequest{
		PageIndex:   o.pageIndex,
		PageSize:    o.pageSize,
		Search:      o.search,
		SortField:   o.sortField,
		SortOrder:   o.sortOrder,
		CreatorID:   o.creatorID,
		BuildStatus: o.buildStatus,
		ByGroup:     o.byGroup,
		GroupPathID: o.groupPathID,
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		q := map[string]interface{}{}
		if req.PageIndex > 0 {
			q["page_index"] = req.PageIndex
		}
		if req.PageSize > 0 {
			q["page_size"] = req.PageSize
		}
		if req.Search != "" {
			q["search"] = req.Search
		}
		if req.SortField != "" {
			q["sort_field"] = req.SortField
		}
		if req.SortOrder != "" {
			q["sort_order"] = req.SortOrder
		}
		if req.CreatorID != "" {
			q["creator_id"] = req.CreatorID
		}
		if req.BuildStatus != "" {
			q["build_status"] = req.BuildStatus
		}
		if req.ByGroup {
			q["by_group"] = true
		}
		if req.GroupPathID != "" {
			q["group_path_id"] = req.GroupPathID
		}
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "GET",
			"path":       fmt.Sprintf("/v1/job/%s/list", o.projectID),
			"project_id": o.projectID,
			"gateway":    cfg.Gateway,
			"query":      q,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListProjectJobs(context.Background(), o.projectID, req)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ------------------------------ build run ------------------------------

type buildRunOpts struct {
	jobID     string
	params    []string // KEY=VAL (repeatable)
	branch    string
	buildTag  string
	commitID  string
	buildType string
	repoID    string
	repoName  string
	scmType   string
	url       string
	webURL    string
	bodyJSON  string
	bodyFile  string
	dryRun    bool
}

func newBuildRunCmd() *cobra.Command {
	o := &buildRunOpts{}
	cmd := &cobra.Command{
		Use:   "run <job_id>",
		Short: "Trigger a build (ExecuteJob API)",
		Long: `Trigger a CodeArts Build job.

<job_id> is the build task ID returned by ` + "`build list`" + ` (32-char string).
It is written into the request body's job_id field.

EXAMPLES:
    # Simplest — use stored job defaults
    codearts-cli build run <job_id>

    # Override branch + build-type + scm
    codearts-cli build run <job_id> \
      --branch main --build-type branch \
      --scm-type codehub --repo-id 8147520

    # Pass build parameters (repeatable or comma-separated)
    codearts-cli build run <job_id> \
      --param "ENV=staging" --param "VERSION=1.2.0"

    # Full body from a file
    codearts-cli build run <job_id> --body-file build.json

API reference: https://support.huaweicloud.com/api-codeci/ExecuteJob.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.jobID = args[0]
			return runBuildRun(cmd, o)
		},
	}
	cmd.Flags().StringArrayVar(&o.params, "param", nil, "Build parameter KEY=VAL (repeatable; comma-separated also works)")
	cmd.Flags().StringVar(&o.branch, "branch", "", "scm.branch")
	cmd.Flags().StringVar(&o.buildTag, "build-tag", "", "scm.build_tag")
	cmd.Flags().StringVar(&o.commitID, "commit-id", "", "scm.build_commit_id")
	cmd.Flags().StringVar(&o.buildType, "build-type", "", "scm.build_type (branch | tag | commitId)")
	cmd.Flags().StringVar(&o.repoID, "repo-id", "", "scm.repo_id")
	cmd.Flags().StringVar(&o.repoName, "repo-name", "", "scm.repo_name")
	cmd.Flags().StringVar(&o.scmType, "scm-type", "", "scm.scm_type (default | codehub)")
	cmd.Flags().StringVar(&o.url, "scm-url", "", "scm.url")
	cmd.Flags().StringVar(&o.webURL, "scm-web-url", "", "scm.web_url")
	cmd.Flags().StringVar(&o.bodyJSON, "body-json", "", "Full JSON body (overrides flag-based fields)")
	cmd.Flags().StringVar(&o.bodyFile, "body-file", "", "Path to a JSON file for the full body")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runBuildRun(cmd *cobra.Command, o *buildRunOpts) error {
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
		// Populate job_id when the caller forgot it — the positional arg is
		// authoritative about which job to trigger.
		if _, ok := m["job_id"]; !ok {
			m["job_id"] = o.jobID
		}
		body = m
	} else {
		req := &client.ExecuteJobRequest{JobID: o.jobID}

		// --param KEY=VAL, repeatable and comma-separated. Matching the
		// --work-item flattening done in repo mr create.
		for _, entry := range o.params {
			for _, kv := range strings.Split(entry, ",") {
				kv = strings.TrimSpace(kv)
				if kv == "" {
					continue
				}
				eq := strings.IndexByte(kv, '=')
				if eq <= 0 {
					return fmt.Errorf("--param %q must be KEY=VAL", kv)
				}
				req.Parameter = append(req.Parameter, client.ExecuteJobParam{
					Name:  strings.TrimSpace(kv[:eq]),
					Value: strings.TrimSpace(kv[eq+1:]),
				})
			}
		}

		if o.branch != "" || o.buildTag != "" || o.commitID != "" ||
			o.buildType != "" || o.repoID != "" || o.repoName != "" ||
			o.scmType != "" || o.url != "" || o.webURL != "" {
			req.SCM = &client.ExecuteJobSCM{
				BuildTag:      o.buildTag,
				BuildCommitID: o.commitID,
				Branch:        o.branch,
				BuildType:     o.buildType,
				RepoID:        o.repoID,
				RepoName:      o.repoName,
				SCMType:       o.scmType,
				URL:           o.url,
				WebURL:        o.webURL,
			}
		}
		body = req
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":  "POST",
			"path":    "/v1/job/execute",
			"gateway": cfg.Gateway,
			"body":    body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ExecuteJob(context.Background(), body)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Build triggered for job %s", o.jobID)
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ------------------------------ build stop ------------------------------

type buildStopOpts struct {
	jobID   string
	buildNo int
	dryRun  bool
}

func newBuildStopCmd() *cobra.Command {
	o := &buildStopOpts{}
	cmd := &cobra.Command{
		Use:   "stop <job_id> <build_no>",
		Short: "Stop a running build (StopTheJob API)",
		Long: `Stop a running build.

<job_id> is the build task's 32-char ID.
<build_no> is the numeric build number (starts at 1, increments every run).
Find the build number in the build-history panel or in the output of a prior
` + "`build run`" + ` (daily_build_number / actual_build_number).

EXAMPLES:
    codearts-cli build stop <job_id> 105

API reference: https://support.huaweicloud.com/api-codeci/StopTheJob.html`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.jobID = args[0]
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				return fmt.Errorf("build_no must be a positive integer (>= 1), got %q", args[1])
			}
			o.buildNo = n
			return runBuildStop(cmd, o)
		},
	}
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runBuildStop(cmd *cobra.Command, o *buildStopOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":  "POST",
			"path":    fmt.Sprintf("/v1/job/%s/stop", o.jobID),
			"job_id":  o.jobID,
			"gateway": cfg.Gateway,
			"body":    map[string]interface{}{"build_no": o.buildNo},
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.StopTheJob(context.Background(), o.jobID, o.buildNo)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Stop requested for job %s build #%d", o.jobID, o.buildNo)
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}
