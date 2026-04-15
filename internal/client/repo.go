package client

import (
	"context"
	"fmt"
)

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
