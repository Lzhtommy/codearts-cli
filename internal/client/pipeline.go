package client

import (
	"context"
	"fmt"
)

// RunPipelineRequest is the JSON body accepted by Huawei Cloud's
// POST /v5/{project_id}/api/pipelines/{pipeline_id}/run endpoint.
//
// Reference: https://support.huaweicloud.com/api-pipeline/RunPipeline.html
//
// All fields are optional on the wire (the pipeline's configured defaults
// apply when omitted), but callers typically want to override at least
// Sources (branch / commit for a specific run).
type RunPipelineRequest struct {
	Sources           []RunSource       `json:"sources,omitempty"`
	Variables         []RunVariable     `json:"variables,omitempty"`
	ChooseJobs        []string          `json:"choose_jobs,omitempty"`
	ChooseStages      []string          `json:"choose_stages,omitempty"`
	Description       string            `json:"description,omitempty"`
	CustomParameters  map[string]string `json:"custom_parameters,omitempty"`
}

// RunSource is a per-source override, typically the branch/tag/commit to
// run the pipeline against.
type RunSource struct {
	Type       string          `json:"type,omitempty"` // "code" | "artifact" | ...
	Parameters SourceParameter `json:"params,omitempty"`
}

// SourceParameter holds the free-form parameters for a source. Huawei's API
// accepts an object of arbitrary string fields (git_url, branch, etc.).
type SourceParameter map[string]string

// RunVariable represents a custom pipeline variable override.
type RunVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RunPipelineResponse is the envelope returned by RunPipeline. The fields
// below cover the documented response; extra fields are preserved in Extra.
type RunPipelineResponse struct {
	PipelineRunID string                 `json:"pipeline_run_id,omitempty"`
	Extra         map[string]interface{} `json:"-"`
}

// RunPipeline triggers a CodeArts pipeline run.
//
// projectID / pipelineID are path parameters; body carries overrides. When
// body is nil, the pipeline runs with its stored defaults (useful for a
// quick smoke trigger).
func (c *Client) RunPipeline(ctx context.Context, projectID, pipelineID string, body *RunPipelineRequest) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if pipelineID == "" {
		return nil, fmt.Errorf("pipelineID is required")
	}
	path := fmt.Sprintf("/v5/%s/api/pipelines/%s/run", projectID, pipelineID)

	// Use a generic map response so forward-compatible fields from the API
	// (e.g. pipeline_run_id) are surfaced verbatim to the caller instead of
	// silently dropped.
	out := map[string]interface{}{}
	// The API rejects POSTs without a JSON body (PARSE_REQUEST_DATA_EXCEPTION),
	// so always send at least an empty object when the caller passed nil.
	var req interface{} = map[string]interface{}{}
	if body != nil {
		req = body
	}
	if err := c.Do(ctx, "POST", c.PipelineEndpoint(), path, nil, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StopPipelineRun stops a running pipeline instance.
//
// Reference: https://support.huaweicloud.com/api-pipeline/StopPipelineRun.html
// Endpoint: POST /v5/{project_id}/api/pipelines/{pipeline_id}/pipeline-runs/{pipeline_run_id}/stop
//
// The API documents "no request body", but Huawei's APIGW parser still
// requires a JSON payload on POST — we send `{}` for the same reason
// RunPipeline does.
func (c *Client) StopPipelineRun(ctx context.Context, projectID, pipelineID, runID string) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if pipelineID == "" {
		return nil, fmt.Errorf("pipelineID is required")
	}
	if runID == "" {
		return nil, fmt.Errorf("pipeline_run_id is required")
	}
	path := fmt.Sprintf("/v5/%s/api/pipelines/%s/pipeline-runs/%s/stop", projectID, pipelineID, runID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.PipelineEndpoint(), path, nil, map[string]interface{}{}, &out); err != nil {
		return nil, err
	}
	return out, nil
}
