package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/autelrobotics/codearts-cli/internal/client"
	"github.com/autelrobotics/codearts-cli/internal/core"
	"github.com/autelrobotics/codearts-cli/internal/output"
)

func newPipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "CodeArts pipeline operations",
	}
	cmd.AddCommand(newPipelineRunCmd())
	cmd.AddCommand(newPipelineStopCmd())
	return cmd
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
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "Override project_id (default from config)")
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
	projectID := cfg.ProjectID
	if o.projectID != "" {
		projectID = o.projectID
	}
	if o.dryRun {
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
	cmd.Flags().StringVar(&o.projectID, "project-id", "", "Override project_id (default from config)")
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

	projectID := cfg.ProjectID
	if o.projectID != "" {
		projectID = o.projectID
	}

	body, err := buildRunBody(o)
	if err != nil {
		return err
	}

	if o.dryRun {
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
	rawBody, err := firstNonEmpty("--body", o.bodyJSON, "--body-file", o.bodyFile)
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

	srcRaw, err := firstNonEmpty("--sources", o.sourcesJSON, "--sources-file", o.sourcesFile)
	if err != nil {
		return nil, err
	}
	if srcRaw != "" {
		if err := json.Unmarshal([]byte(srcRaw), &req.Sources); err != nil {
			return nil, fmt.Errorf("parse --sources JSON: %w", err)
		}
	}

	varRaw, err := firstNonEmpty("--variables", o.varsJSON, "--variables-file", o.varsFile)
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

// firstNonEmpty returns the value of the first non-empty input (inline JSON
// or file). It errors when both are supplied simultaneously for the same
// logical input.
func firstNonEmpty(inlineName, inline, fileName, file string) (string, error) {
	if inline != "" && file != "" {
		return "", fmt.Errorf("%s and %s are mutually exclusive", inlineName, fileName)
	}
	if inline != "" {
		return inline, nil
	}
	if file != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", fileName, err)
		}
		s := strings.TrimSpace(string(b))
		if s == "" {
			return "", errors.New(fileName + " is empty")
		}
		return s, nil
	}
	return "", nil
}
