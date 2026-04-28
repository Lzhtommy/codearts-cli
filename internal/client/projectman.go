package client

import (
	"context"
	"fmt"
	"net/url"
)

// ----- ListIpdProjectIssues -----
//
// Reference: https://support.huaweicloud.com/api-projectman/ListIpdProjectIssues.html
// Endpoint:  POST /v1/ipdprojectservice/projects/{project_id}/issues/query?issue_type=...
//
// issue_type is a **required** query parameter, comma-separated for multiple
// types (RR/SF/IR/US/Task/Bug/Epic/FE/SR/AR depending on project kind).

// ListIssuesRequest is the POST body for the Huawei ListIpdProjectIssues API.
// All fields are optional — an empty body returns every issue of the given
// issue_type(s) paginated at the API's default page size.
type ListIssuesRequest struct {
	Filter     []map[string]interface{} `json:"filter,omitempty"`
	FilterMode string                   `json:"filter_mode,omitempty"` // OR_AND | AND_OR (default AND_OR)
	Page       *PageInfo                `json:"page,omitempty"`
	Sort       []SortInfo               `json:"sort,omitempty"`
}

type PageInfo struct {
	PageNo   int `json:"page_no,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

type SortInfo struct {
	Field string `json:"field"`
	Asc   bool   `json:"asc"`
}

// ListIpdProjectIssues queries work items in a project.
func (c *Client) ListIpdProjectIssues(ctx context.Context, projectID, issueType string, body *ListIssuesRequest) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if issueType == "" {
		return nil, fmt.Errorf("issue_type is required (e.g. US, Task, Bug; comma-separated for multiple)")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/query", projectID)
	q := url.Values{"issue_type": {issueType}}
	// Empty body → send {} so APIGW's JSON parser is happy (same reason as
	// RunPipeline — see pipeline.go). body may be nil.
	var req interface{} = map[string]interface{}{}
	if body != nil {
		req = body
	}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.ProjectManEndpoint(), path, q, req, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- ShowIssueDetail -----
//
// Reference: https://support.huaweicloud.com/api-projectman/ShowIssueDetail.html
// Endpoint:  GET /v1/ipdprojectservice/projects/{project_id}/issues/{issue_id}?issue_type=...&domain_id=...

// ShowIssueDetail fetches a single work item by ID. issue_type is required
// (one of Epic/FE/SF/IR/RR/SR/US/AR/Bug/Task). domainID is optional.
func (c *Client) ShowIssueDetail(ctx context.Context, projectID, issueID, issueType, domainID string) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if issueID == "" {
		return nil, fmt.Errorf("issue_id is required")
	}
	if issueType == "" {
		return nil, fmt.Errorf("issue_type is required (e.g. US, Task, Bug)")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/%s", projectID, issueID)
	q := url.Values{"issue_type": {issueType}}
	if domainID != "" {
		q.Set("domain_id", domainID)
	}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.ProjectManEndpoint(), path, q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- CreateIpdProjectIssue -----
//
// Reference: https://support.huaweicloud.com/api-projectman/CreateIpdProjectIssue.html
// Endpoint:  POST /v1/ipdprojectservice/projects/{project_id}/issues

// CreateIssueRequest models the minimum-plus-common fields for creating an
// issue. The API accepts many more optional fields (business_domain,
// plan_iteration, workload_man_day, …); for those, use CreateIssueRaw with
// a free-form map instead.
type CreateIssueRequest struct {
	Title       string `json:"title"`               // required, ≤ 256 chars
	Description string `json:"description"`         // required, ≤ 500000 chars
	Category    string `json:"category"`            // required: RR|SF|IR|SR|AR|Task|Bug|US|Epic|FE
	Assignee    string `json:"assignee"`            // required: 32-char user_id UUID
	Status      string `json:"status,omitempty"`    // optional
	Priority    string `json:"priority,omitempty"`  // optional
}

// CreateIpdProjectIssue creates a new work item. Passing a nil body returns
// an error since all four fields are required; the structured type prevents
// accidentally omitting them.
func (c *Client) CreateIpdProjectIssue(ctx context.Context, projectID string, body interface{}) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if body == nil {
		return nil, fmt.Errorf("request body is required (title/description/category/assignee)")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues", projectID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.ProjectManEndpoint(), path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- ListE2EGraphsOpenAPI -----
//
// Reference: https://support.huaweicloud.com/api-projectman/ListE2EGraphsOpenAPI.html
// Endpoint:  GET /v1/ipdprojectservice/projects/{project_id}/e2e/graphs?issue_id=&category=&is_src=
//
// Returns the end-to-end trace graph (trace_list) for a single work item —
// parent/child issues, associated commits/MRs, branches, testcases, etc.

// ListE2EGraphs fetches the E2E trace graph for one work item.
// issueID must match the API regex (18–19 digits); category is one of
// RR/SF/IR/SR/AR/Task/Bug/Epic/FE/US. isSrc is a tri-state pointer: pass
// nil to omit, or a bool pointer to explicitly include is_src=true/false
// for cross-project queries.
func (c *Client) ListE2EGraphs(ctx context.Context, projectID, issueID, category string, isSrc *bool) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if issueID == "" {
		return nil, fmt.Errorf("issue_id is required")
	}
	if category == "" {
		return nil, fmt.Errorf("category is required (e.g. US, Task, Bug)")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/e2e/graphs", projectID)
	q := url.Values{
		"issue_id": {issueID},
		"category": {category},
	}
	if isSrc != nil {
		if *isSrc {
			q.Set("is_src", "true")
		} else {
			q.Set("is_src", "false")
		}
	}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.ProjectManEndpoint(), path, q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- ListProjectUsers -----
//
// Reference: https://support.huaweicloud.com/api-projectman/ListProjectUsers.html
// Endpoint:  GET /v1/ipdprojectservice/projects/{project_id}/users
//
// Returns the member list (user_id, user_name, nick_name, role_name, …) for
// a project. No query params.

// ListProjectUsers queries project members.
func (c *Client) ListProjectUsers(ctx context.Context, projectID string) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/users", projectID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.ProjectManEndpoint(), path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- ListIssueStatues -----
//
// Reference: https://support.huaweicloud.com/api-projectman/ListIssueStatues.html
// Endpoint:  GET /v1/ipdprojectservice/projects/{project_id}/category/{category_id}/statuses
//
// Returns the status definitions for a given work-item type (category_id is
// the 5-digit numeric type ID, e.g. 10001/10020/10021/10022/10023/10027/
// 10028/10029/10033/10065 — NOT the RR/Bug/Task string category).

// ListIssueStatues returns the status definitions for a work-item type.
func (c *Client) ListIssueStatues(ctx context.Context, projectID, categoryID string) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if categoryID == "" {
		return nil, fmt.Errorf("category_id is required (5-digit numeric work-item type ID)")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/category/%s/statuses", projectID, categoryID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.ProjectManEndpoint(), path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- BatchUpdateIpdIssues -----
//
// Reference: https://support.huaweicloud.com/api-projectman/BatchUpdateIpdIssues.html
// Endpoint:  PUT /v1/ipdprojectservice/projects/{project_id}/issues/batch

// BatchUpdateIssuesRequest updates `attribute` on a list of issues by ID.
type BatchUpdateIssuesRequest struct {
	IDs       []string               `json:"id"`        // required: issue IDs to update
	Attribute map[string]interface{} `json:"attribute"` // required: must include `category`
}

// BatchUpdateIpdIssues applies the same attribute changes to many issues.
func (c *Client) BatchUpdateIpdIssues(ctx context.Context, projectID string, body *BatchUpdateIssuesRequest) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if body == nil || len(body.IDs) == 0 {
		return nil, fmt.Errorf("at least one issue id is required")
	}
	if body.Attribute == nil || body.Attribute["category"] == nil || body.Attribute["category"] == "" {
		return nil, fmt.Errorf("attribute.category is required")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/batch", projectID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "PUT", c.ProjectManEndpoint(), path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- CreateIssueComment -----
//
// No public Huawei OpenAPI doc — observed in CodeArts UI. Verified against
// projectman-ext.cn-south-1 with AK/SK signing (see tools/probe-comments).
// Endpoint: POST /v1/ipdprojectservice/projects/{project_id}/issues/{issue_id}/comments

// CreateIssueCommentRequest is the body for adding a comment to a work item.
// `category` is the entity-type discriminator (always "comment"); `issue_category`
// is the work-item type the comment belongs to (Task/Bug/US/...). `description`
// is HTML — the UI uses a rich-text editor and stores `<p>...</p>` markup.
type CreateIssueCommentRequest struct {
	Category      string `json:"category"`       // always "comment"
	IssueCategory string `json:"issue_category"` // Task|Bug|US|RR|SF|IR|SR|AR|Epic|FE
	Description   string `json:"description"`    // HTML body
}

// CreateIssueComment posts a comment to a work item.
func (c *Client) CreateIssueComment(ctx context.Context, projectID, issueID string, body interface{}) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if issueID == "" {
		return nil, fmt.Errorf("issue_id is required")
	}
	if body == nil {
		return nil, fmt.Errorf("request body is required (issue_category/description)")
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/%s/comments", projectID, issueID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.ProjectManEndpoint(), path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ----- ListIssueComments -----
//
// No public Huawei OpenAPI doc — observed in CodeArts UI and verified
// end-to-end against projectman-ext.cn-south-1 with AK/SK signing.
// Endpoint: GET /v1/ipdprojectservice/projects/{project_id}/issues/{issue_id}/comments
//
// `category` is **required** by the upstream (omitting it triggers a 200 with
// `status:"failed"` and message "Required request paramteter 'category' can
// not be null"). Valid values: any subset of `comment,reply,operation`,
// comma-separated. The response shape is `result.comment_list[]` where each
// entry has id / issue_id / category (Comment|Reply|Operation) / description
// (HTML) / creator_info / created_date (ms) / top / top_flag, plus
// extend_attribute_obj for Operation entries.

// ListIssueCommentsOptions are the query parameters for ListIssueComments.
// Category defaults to "comment,reply,operation" if empty (matches the UI).
// PageNo/PageSize default to the upstream's defaults when 0.
// TargetProjectID is used for cross-project queries; empty means same project.
// DateDesc nil omits the param (upstream default applies).
type ListIssueCommentsOptions struct {
	Category        string
	PageNo          int
	PageSize        int
	DateDesc        *bool
	TargetProjectID string
}

// ListIssueComments fetches the comment / reply / operation log entries for a
// work item. opts may be nil; in that case category defaults to all three
// types and pagination uses upstream defaults.
func (c *Client) ListIssueComments(ctx context.Context, projectID, issueID string, opts *ListIssueCommentsOptions) (map[string]interface{}, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}
	if issueID == "" {
		return nil, fmt.Errorf("issue_id is required")
	}
	if opts == nil {
		opts = &ListIssueCommentsOptions{}
	}
	category := opts.Category
	if category == "" {
		category = "comment,reply,operation"
	}
	q := url.Values{"category": {category}}
	if opts.PageNo > 0 {
		q.Set("page_no", fmt.Sprintf("%d", opts.PageNo))
	}
	if opts.PageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", opts.PageSize))
	}
	if opts.DateDesc != nil {
		if *opts.DateDesc {
			q.Set("date_desc", "true")
		} else {
			q.Set("date_desc", "false")
		}
	}
	if opts.TargetProjectID != "" {
		q.Set("target_project_id", opts.TargetProjectID)
	}
	path := fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/%s/comments", projectID, issueID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.ProjectManEndpoint(), path, q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
