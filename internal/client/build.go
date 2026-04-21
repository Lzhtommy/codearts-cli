package client

import (
	"context"
	"fmt"
	"net/url"
)

// ---------- ListProjectJobs ----------
//
// Reference: https://support.huaweicloud.com/api-codeci/ListProjectJobs.html
// Endpoint:  GET /v1/job/{project_id}/list

// ListProjectJobsRequest bundles the query parameters for ListProjectJobs.
// All fields are optional — the API applies defaults (page_index=0,
// page_size=10) when unset.
type ListProjectJobsRequest struct {
	PageIndex   int    // 0..999999999, default 0
	PageSize    int    // 1..100, default 10
	Search      string // fuzzy query on job name / creator
	SortField   string
	SortOrder   string
	CreatorID   string
	BuildStatus string // red | blue | timeout | aborted | building | none
	ByGroup     bool   // enable grouping
	GroupPathID string
}

// ListProjectJobs queries build jobs in a project.
func (c *Client) ListProjectJobs(ctx context.Context, projectID string, req *ListProjectJobsRequest) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	path := fmt.Sprintf("/v1/job/%s/list", projectID)
	q := url.Values{}
	if req != nil {
		if req.PageIndex > 0 {
			q.Set("page_index", fmt.Sprintf("%d", req.PageIndex))
		}
		if req.PageSize > 0 {
			q.Set("page_size", fmt.Sprintf("%d", req.PageSize))
		}
		if req.Search != "" {
			q.Set("search", req.Search)
		}
		if req.SortField != "" {
			q.Set("sort_field", req.SortField)
		}
		if req.SortOrder != "" {
			q.Set("sort_order", req.SortOrder)
		}
		if req.CreatorID != "" {
			q.Set("creator_id", req.CreatorID)
		}
		if req.BuildStatus != "" {
			q.Set("build_status", req.BuildStatus)
		}
		if req.ByGroup {
			q.Set("by_group", "true")
		}
		if req.GroupPathID != "" {
			q.Set("group_path_id", req.GroupPathID)
		}
	}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.BuildEndpoint(), path, q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- ExecuteJob ----------
//
// Reference: https://support.huaweicloud.com/api-codeci/ExecuteJob.html
// Endpoint:  POST /v1/job/execute

// ExecuteJobRequest is the JSON body for ExecuteJob.
//
// JobID is technically optional in the wire spec (Huawei's docs phrase it
// as "建议传"), but practically every trigger path we care about targets a
// specific job — the CLI enforces it.
type ExecuteJobRequest struct {
	JobID     string            `json:"job_id,omitempty"`
	Parameter []ExecuteJobParam `json:"parameter,omitempty"`
	SCM       *ExecuteJobSCM    `json:"scm,omitempty"`
}

// ExecuteJobParam is a single {name,value} build parameter pair.
type ExecuteJobParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExecuteJobSCM is the optional source-code override block. All fields are
// optional — populate only what you need to override on this run.
type ExecuteJobSCM struct {
	BuildTag      string `json:"build_tag,omitempty"`
	BuildCommitID string `json:"build_commit_id,omitempty"`
	Branch        string `json:"branch,omitempty"`
	BuildType     string `json:"build_type,omitempty"` // branch | tag | commitId
	RepoID        string `json:"repo_id,omitempty"`
	RepoName      string `json:"repo_name,omitempty"`
	SCMType       string `json:"scm_type,omitempty"` // default | codehub
	URL           string `json:"url,omitempty"`
	WebURL        string `json:"web_url,omitempty"`
}

// ExecuteJob triggers a build job.
//
// The API accepts `body` as either a typed *ExecuteJobRequest or a free-form
// map (for fields we haven't modeled). A non-nil JSON payload must always be
// sent — Huawei's APIGW rejects empty POSTs with PARSE_REQUEST_DATA_EXCEPTION.
func (c *Client) ExecuteJob(ctx context.Context, body interface{}) (map[string]interface{}, error) {
	if body == nil {
		body = map[string]interface{}{}
	}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.BuildEndpoint(), "/v1/job/execute", nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- StopTheJob ----------
//
// Reference: https://support.huaweicloud.com/api-codeci/StopTheJob.html
// Endpoint:  POST /v1/job/{job_id}/stop
// Body:      {"build_no": <int>} — required; >= 1.

// StopTheJob stops a running build. buildNo is the numeric build number
// (starts at 1, increments every run) that identifies which run to stop.
func (c *Client) StopTheJob(ctx context.Context, jobID string, buildNo int) (map[string]interface{}, error) {
	if jobID == "" {
		return nil, fmt.Errorf("jobID is required")
	}
	if buildNo < 1 {
		return nil, fmt.Errorf("build_no must be >= 1 (the build number to stop)")
	}
	path := fmt.Sprintf("/v1/job/%s/stop", jobID)
	body := map[string]interface{}{"build_no": buildNo}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.BuildEndpoint(), path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
