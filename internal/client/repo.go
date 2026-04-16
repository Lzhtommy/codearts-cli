package client

import (
	"context"
	"fmt"
	"net/url"
)

// ListRepositories queries repositories in a project.
//
// Reference: https://support.huaweicloud.com/api-codeartsrepo/ShowAllRepositoryByTwoProjectId.html
// Endpoint:  GET /v2/projects/{project_uuid}/repositories?page_index=&page_size=&search=
func (c *Client) ListRepositories(ctx context.Context, projectUUID string, pageIndex, pageSize int, search string) (map[string]interface{}, error) {
	if projectUUID == "" {
		return nil, fmt.Errorf("project_uuid is required")
	}
	path := fmt.Sprintf("/v2/projects/%s/repositories", projectUUID)
	q := url.Values{}
	if pageIndex > 0 {
		q.Set("page_index", fmt.Sprintf("%d", pageIndex))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	if search != "" {
		q.Set("search", search)
	}
	out := map[string]interface{}{}
	if err := c.Do(ctx, "GET", c.RepoEndpoint(), path, q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateMergeRequest creates a merge request on a repository.
//
// Reference: https://support.huaweicloud.com/api-codeartsrepo/CreateMergeRequest.html
// Endpoint:  POST /v4/repositories/{repository_id}/merge-requests
//
// title, source_branch and target_branch are the minimum required body
// fields. All other fields (reviewer_ids, labels, squash, work_item_ids,
// …) are optional — pass them via a free-form map or the struct below.

// CreateMRRequest models the required+common fields for creating an MR.
// For the long-tail optional fields (labels, milestone_id, squash,
// work_item_ids, …), pass a free-form map to CreateMergeRequest instead.
type CreateMRRequest struct {
	Title                   string   `json:"title"`                                // required
	SourceBranch            string   `json:"source_branch"`                        // required
	TargetBranch            string   `json:"target_branch"`                        // required
	Description             string   `json:"description,omitempty"`
	TargetRepositoryID      int      `json:"target_repository_id,omitempty"`
	ReviewerIDs             string   `json:"reviewer_ids,omitempty"`               // comma-separated
	AssigneeIDs             string   `json:"assignee_ids,omitempty"`               // comma-separated
	ApprovalReviewerIDs     string   `json:"approval_reviewer_ids,omitempty"`      // comma-separated
	ApprovalApproversIDs    string   `json:"approval_approvers_ids,omitempty"`     // comma-separated
	MilestoneID             int      `json:"milestone_id,omitempty"`
	ForceRemoveSourceBranch bool     `json:"force_remove_source_branch,omitempty"`
	Squash                  bool     `json:"squash,omitempty"`
	SquashCommitMessage     string   `json:"squash_commit_message,omitempty"`
	WorkItemIDs             []string `json:"work_item_ids,omitempty"`
	IsUseTempBranch         bool     `json:"is_use_temp_branch,omitempty"`
	OnlyAssigneeCanMerge    bool     `json:"only_assignee_can_merge,omitempty"`
}

// CreateMergeRequest posts a new MR. body may be *CreateMRRequest or a
// free-form map when extra fields (labels, …) are needed.
func (c *Client) CreateMergeRequest(ctx context.Context, repositoryID int, body interface{}) (map[string]interface{}, error) {
	if repositoryID <= 0 {
		return nil, fmt.Errorf("repository_id must be a positive integer")
	}
	if body == nil {
		return nil, fmt.Errorf("request body is required (title/source_branch/target_branch)")
	}
	if req, ok := body.(*CreateMRRequest); ok {
		if req.Title == "" || req.SourceBranch == "" || req.TargetBranch == "" {
			return nil, fmt.Errorf("title, source_branch and target_branch are all required")
		}
	}
	path := fmt.Sprintf("/v4/repositories/%d/merge-requests", repositoryID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.RepoEndpoint(), path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateMergeRequestDiscussion creates a review discussion on a merge
// request.
//
// Reference: https://support.huaweicloud.com/api-codeartsrepo/CreateMergeRequestDiscussion.html
// Endpoint:  POST /v4/repositories/{repository_id}/merge-requests/{merge_request_iid}/discussions
//
// Note: repositoryID and mergeRequestIID are numeric in the API. We accept
// them as int so callers can't accidentally pass a UUID-shaped project id.

// CreateMRDiscussionRequest models a review comment on a MR.
//
// Only `body` is required. The richer review-metadata fields (severity,
// review_categories, review_modules, proposer_id, line_types, position)
// are optional — use CreateMRDiscussionRaw with a free-form map if you need
// to post a line-level code comment (`position` has a nested structure).
type CreateMRDiscussionRequest struct {
	Body             string `json:"body"`
	Severity         string `json:"severity,omitempty"`          // suggestion | minor | major | fatal
	AssigneeID       string `json:"assignee_id,omitempty"`
	ReviewCategories string `json:"review_categories,omitempty"`
	ReviewModules    string `json:"review_modules,omitempty"`
	ProposerID       string `json:"proposer_id,omitempty"`
	LineTypes        string `json:"line_types,omitempty"`
}

// CreateMergeRequestDiscussion posts a new discussion.
// When body is a *CreateMRDiscussionRequest the Body field must be non-empty.
// When body is a free-form map the caller is responsible for required fields.
func (c *Client) CreateMergeRequestDiscussion(ctx context.Context, repositoryID, mergeRequestIID int, body interface{}) (map[string]interface{}, error) {
	if repositoryID <= 0 {
		return nil, fmt.Errorf("repository_id must be a positive integer")
	}
	if mergeRequestIID <= 0 {
		return nil, fmt.Errorf("merge_request_iid must be a positive integer")
	}
	if body == nil {
		return nil, fmt.Errorf("request body is required (at minimum a `body` field)")
	}
	if req, ok := body.(*CreateMRDiscussionRequest); ok && req.Body == "" {
		return nil, fmt.Errorf("`body` (comment content) is required")
	}
	path := fmt.Sprintf("/v4/repositories/%d/merge-requests/%d/discussions", repositoryID, mergeRequestIID)
	out := map[string]interface{}{}
	if err := c.Do(ctx, "POST", c.RepoEndpoint(), path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
