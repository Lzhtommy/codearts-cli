package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Lzhtommy/codearts-cli/internal/client"
	"github.com/Lzhtommy/codearts-cli/internal/core"
	"github.com/Lzhtommy/codearts-cli/internal/output"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "CodeArts pipeline operations",
	}
	cmd.AddCommand(newPipelineListCmd())
	cmd.AddCommand(newPipelineRunCmd())
	cmd.AddCommand(newPipelineStopCmd())
	return cmd
}

// ------------------------------ pipeline list ------------------------------

type pipelineListOpts struct {
	projectID   string
	name        string
	status      []string
	creatorIDs  []string
	executorIDs []string
	startTime   string
	endTime     string
	offset      int
	limit       int
	sortKey     string
	sortDir     string
	dryRun      bool
}

func newPipelineListCmd() *cobra.Command {
	o := &pipelineListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines in a project (ListPipelines API)",
		Long: `List pipelines in a CodeArts project.

--project-id is required. It is used both as the URL path parameter and
automatically injected into the body's project_ids array, per the API spec.

EXAMPLES:
    # List all pipelines
    codearts-cli pipeline list --project-id <proj>

    # Filter by name
    codearts-cli pipeline list --project-id <proj> --name "deploy"

    # Paginate
    codearts-cli pipeline list --project-id <proj> --offset 0 --limit 20

API reference: https://support.huaweicloud.com/api-pipeline/ListPipelines.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPipelineList(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "(required) Huawei Cloud project_id")
	cmd.Flags().StringVar(&o.name, "name", "", "Filter by pipeline name (fuzzy match)")
	cmd.Flags().StringSliceVar(&o.status, "status", nil, "Filter by status (repeatable): COMPLETED | RUNNING | FAILED | CANCELED | PAUSED | SUSPEND | IGNORED")
	cmd.Flags().StringSliceVar(&o.creatorIDs, "creator-id", nil, "Filter by creator user_id (repeatable)")
	cmd.Flags().StringSliceVar(&o.executorIDs, "executor-id", nil, "Filter by executor user_id (repeatable)")
	cmd.Flags().StringVar(&o.startTime, "start-time", "", "Filter: created after this time")
	cmd.Flags().StringVar(&o.endTime, "end-time", "", "Filter: created before this time")
	cmd.Flags().IntVar(&o.offset, "offset", 0, "Pagination offset (default 0)")
	cmd.Flags().IntVar(&o.limit, "limit", 0, "Pagination limit (0 = API default)")
	cmd.Flags().StringVar(&o.sortKey, "sort-key", "", "Sort field: name | create_time | update_time")
	cmd.Flags().StringVar(&o.sortDir, "sort-dir", "", "Sort direction: asc | desc")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runPipelineList(cmd *cobra.Command, o *pipelineListOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := o.projectID
	if projectID == "" {
		return fmt.Errorf("--project-id is required for pipeline commands")
	}

	body := &client.ListPipelinesRequest{
		// Per API spec and user requirement: project_id in path AND in body.
		ProjectID:   projectID,
		ProjectIDs:  []string{projectID},
		Name:        o.name,
		Status:      o.status,
		CreatorIDs:  o.creatorIDs,
		ExecutorIDs: o.executorIDs,
		StartTime:   o.startTime,
		EndTime:     o.endTime,
		Offset:      o.offset,
		Limit:       o.limit,
		SortKey:     o.sortKey,
		SortDir:     o.sortDir,
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "POST",
			"project_id": projectID,
			"path":       fmt.Sprintf("/v5/%s/api/pipelines/list", projectID),
			"region":     cfg.Region,
			"body":       body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListPipelines(context.Background(), projectID, body)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ------------------------------ pipeline stop ------------------------------

type pipelineStopOpts struct {
	projectID  string
	pipelineID string
	runID      string
	dryRun     bool
}

func newPipelineStopCmd() *cobra.Command {
	o := &pipelineStopOpts{}
	cmd := &cobra.Command{
		Use:   "stop <pipeline_id> <pipeline_run_id>",
		Short: "Stop a running pipeline instance (StopPipelineRun API)",
		Long: `Stop a running pipeline instance.

Both pipeline_id and pipeline_run_id are required. The latter comes from the
RunPipeline response (or the "last run" field on the pipeline detail page).

API reference: https://support.huaweicloud.com/api-pipeline/StopPipelineRun.html`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.pipelineID = args[0]
			o.runID = args[1]
			return runPipelineStop(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "(required) Huawei Cloud project_id")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit without calling the API")
	return cmd
}

func runPipelineStop(cmd *cobra.Command, o *pipelineStopOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := o.projectID
	if projectID == "" {
		return fmt.Errorf("--project-id is required for pipeline commands")
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":          "POST",
			"project_id":      projectID,
			"pipeline_id":     o.pipelineID,
			"pipeline_run_id": o.runID,
			"path":            fmt.Sprintf("/v5/%s/api/pipelines/%s/pipeline-runs/%s/stop", projectID, o.pipelineID, o.runID),
			"region":          cfg.Region,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.StopPipelineRun(context.Background(), projectID, o.pipelineID, o.runID)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Pipeline run %s stop requested", o.runID)
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

type pipelineRunOpts struct {
	projectID    string
	pipelineID   string
	sourcesJSON  string
	sourcesFile  string
	varsJSON     string
	varsFile     string
	bodyJSON     string
	bodyFile     string
	description  string
	chooseJobs   []string
	chooseStages []string
	dryRun       bool
}

func newPipelineRunCmd() *cobra.Command {
	o := &pipelineRunOpts{}
	cmd := &cobra.Command{
		Use:   "run <pipeline_id>",
		Short: "Trigger a pipeline run (RunPipeline API)",
		Long: `Trigger a CodeArts pipeline run.

The pipeline_id is required (positional arg). project_id and region come from
your config unless overridden with flags.

EXAMPLES:
    # Simplest: run with pipeline defaults
    codearts-cli pipeline run 7f3a...

    # Override the source branch
    codearts-cli pipeline run 7f3a... \
        --sources '[{"type":"code","params":{"build_type":"branch","target_branch":"main"}}]'

    # Pass custom variables
    codearts-cli pipeline run 7f3a... \
        --variables '[{"name":"ENV","value":"staging"}]'

    # Full body from a file (covers every option the API accepts)
    codearts-cli pipeline run 7f3a... --body-file run.json

    # Preview the signed request without sending it
    codearts-cli pipeline run 7f3a... --dry-run

API reference: https://support.huaweicloud.com/api-pipeline/RunPipeline.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.pipelineID = args[0]
			return runPipelineRun(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "(required) Huawei Cloud project_id")
	cmd.Flags().StringVar(&o.sourcesJSON, "sources", "", "JSON array of source overrides")
	cmd.Flags().StringVar(&o.sourcesFile, "sources-file", "", "Path to a JSON file containing the sources array")
	cmd.Flags().StringVar(&o.varsJSON, "variables", "", "JSON array of {name,value} variable overrides")
	cmd.Flags().StringVar(&o.varsFile, "variables-file", "", "Path to a JSON file containing the variables array")
	cmd.Flags().StringVar(&o.bodyJSON, "body", "", "Full request body as JSON — overrides --sources/--variables")
	cmd.Flags().StringVar(&o.bodyFile, "body-file", "", "Path to a JSON file containing the full request body")
	cmd.Flags().StringVar(&o.description, "description", "", "Human-readable description stored on the run")
	cmd.Flags().StringSliceVar(&o.chooseJobs, "choose-job", nil, "Restrict run to specific job IDs (repeatable)")
	cmd.Flags().StringSliceVar(&o.chooseStages, "choose-stage", nil, "Restrict run to specific stage IDs (repeatable)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request body and exit without calling the API")
	return cmd
}

func runPipelineRun(cmd *cobra.Command, o *pipelineRunOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	projectID := o.projectID
	if projectID == "" {
		return fmt.Errorf("--project-id is required for pipeline commands")
	}

	body, err := buildRunBody(o)
	if err != nil {
		return err
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":      "POST",
			"project_id":  projectID,
			"pipeline_id": o.pipelineID,
			"path":        fmt.Sprintf("/v5/%s/api/pipelines/%s/run", projectID, o.pipelineID),
			"region":      cfg.Region,
			"body":        body,
		})
		return nil
	}

	cli, err := client.New(cfg)
	if err != nil {
		return err
	}

	resp, err := cli.RunPipeline(context.Background(), projectID, o.pipelineID, body)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Pipeline %s triggered in project %s", o.pipelineID, projectID)
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// buildRunBody assembles a RunPipelineRequest from the flag combinations.
//
// Flag precedence (last wins): explicit --body / --body-file >
// component flags (--sources, --variables, --description, --choose-job,
// --choose-stage). When --body is provided, component flags are ignored to
// avoid ambiguous merges.
func buildRunBody(o *pipelineRunOpts) (*client.RunPipelineRequest, error) {
	// Full-body path
	rawBody, err := FirstNonEmpty("--body", o.bodyJSON, "--body-file", o.bodyFile)
	if err != nil {
		return nil, err
	}
	if rawBody != "" {
		req := &client.RunPipelineRequest{}
		if err := json.Unmarshal([]byte(rawBody), req); err != nil {
			return nil, fmt.Errorf("parse body JSON: %w", err)
		}
		return req, nil
	}

	req := &client.RunPipelineRequest{
		Description:  o.description,
		ChooseJobs:   o.chooseJobs,
		ChooseStages: o.chooseStages,
	}

	srcRaw, err := FirstNonEmpty("--sources", o.sourcesJSON, "--sources-file", o.sourcesFile)
	if err != nil {
		return nil, err
	}
	if srcRaw != "" {
		if err := json.Unmarshal([]byte(srcRaw), &req.Sources); err != nil {
			return nil, fmt.Errorf("parse --sources JSON: %w", err)
		}
	}

	varRaw, err := FirstNonEmpty("--variables", o.varsJSON, "--variables-file", o.varsFile)
	if err != nil {
		return nil, err
	}
	if varRaw != "" {
		if err := json.Unmarshal([]byte(varRaw), &req.Variables); err != nil {
			return nil, fmt.Errorf("parse --variables JSON: %w", err)
		}
	}

	// If the caller passed no body overrides at all, let the API take
	// stored pipeline defaults by sending nil.
	if len(req.Sources) == 0 && len(req.Variables) == 0 &&
		len(req.ChooseJobs) == 0 && len(req.ChooseStages) == 0 &&
		req.Description == "" {
		return nil, nil
	}
	return req, nil
}

// FirstNonEmpty returns the value of the first non-empty input (inline JSON
// or file). It errors when both are supplied simultaneously for the same
// logical input.
func FirstNonEmpty(inlineName, inline, fileName, file string) (string, error) {
	if inline != "" && file != "" {
		return "", fmt.Errorf("%s and %s are mutually exclusive — pass only one", inlineName, fileName)
	}
	if inline != "" {
		return inline, nil
	}
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read file %q (via %s): %w", file, fileName, err)
		}
		s := strings.TrimSpace(string(b))
		if s == "" {
			return "", fmt.Errorf("file %q (via %s) is empty", file, fileName)
		}
		return s, nil
	}
	return "", nil
}
