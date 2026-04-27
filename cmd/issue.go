package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Lzhtommy/codearts-cli/internal/client"
	"github.com/Lzhtommy/codearts-cli/internal/core"
	"github.com/Lzhtommy/codearts-cli/internal/output"
)

func newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "CodeArts ProjectMan work-item (IPD) operations",
	}
	cmd.AddCommand(newIssueListCmd())
	cmd.AddCommand(newIssueShowCmd())
	cmd.AddCommand(newIssueCreateCmd())
	cmd.AddCommand(newIssueBatchUpdateCmd())
	cmd.AddCommand(newIssueRelationsCmd())
	cmd.AddCommand(newIssueMembersCmd())
	cmd.AddCommand(newIssueStatusesCmd())
	cmd.AddCommand(newIssueCommentCmd())
	return cmd
}

// ----------------------- issue statuses -----------------------

type issueStatusesOpts struct {
	categoryID string
	dryRun     bool
}

func newIssueStatusesCmd() *cobra.Command {
	o := &issueStatusesOpts{}
	cmd := &cobra.Command{
		Use:   "statuses <category_id>",
		Short: "List status definitions for a work-item type (ListIssueStatues)",
		Long: `List the status definitions configured on a work-item type in the
current project.

<category_id> is the 5-digit **numeric** work-item type ID (NOT the
RR/Bug/Task string). Valid IDs per the API:
    10001, 10020, 10021, 10022, 10023, 10027, 10028, 10029, 10033, 10065
The exact name→id mapping is project-specific — look it up in the CodeArts
Req console (工作项类型 settings) or in the API response of a prior query.

Each returned status includes:
  - name       — human-readable status label
  - belonging  — lifecycle bucket: START | IN_PROGRESS | END

EXAMPLES:
    codearts-cli issue statuses 10020          # statuses of type 10020
    codearts-cli issue statuses 10001 --dry-run

API reference: https://support.huaweicloud.com/api-projectman/ListIssueStatues.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.categoryID = args[0]
			return runIssueStatuses(cmd, o)
		},
	}
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runIssueStatuses(cmd *cobra.Command, o *issueStatusesOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}
	if !isFiveDigit(o.categoryID) {
		return fmt.Errorf("category_id must be a 5-digit numeric ID (e.g. 10020), got %q", o.categoryID)
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":      "GET",
			"path":        fmt.Sprintf("/v1/ipdprojectservice/projects/%s/category/%s/statuses", projectID, o.categoryID),
			"project_id":  projectID,
			"category_id": o.categoryID,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListIssueStatues(context.Background(), projectID, o.categoryID)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// isFiveDigit reports whether s is exactly five ASCII digits. Keeps the
// regex-in-the-API-spec check local and dependency-free.
func isFiveDigit(s string) bool {
	if len(s) != 5 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// ----------------------- issue relations -----------------------

type issueRelationsOpts struct {
	issueID  string
	category string
	isSrc    string // "" | "true" | "false" — tri-state passthrough
	dryRun   bool
}

func newIssueRelationsCmd() *cobra.Command {
	o := &issueRelationsOpts{}
	cmd := &cobra.Command{
		Use:   "relations <issue_id>",
		Short: "Query E2E trace graph for a work item (ListE2EGraphsOpenAPI)",
		Long: `Return the end-to-end trace graph for a work item: parent/child
issues, associated work items, commits, MRs, branches, testcases, testplans,
and documents.

<issue_id> must be the 18-19 digit numeric issue ID, not the short number
visible in the console (that is ` + "`number`" + ` — use the API response ` + "`id`" + ` field).

EXAMPLES:
    # Traces for a User Story
    codearts-cli issue relations 1251275102548402177 --category US

    # Cross-project query (src = upstream / dst = downstream)
    codearts-cli issue relations 1251275102548402177 --category Bug --is-src true

API reference: https://support.huaweicloud.com/api-projectman/ListE2EGraphsOpenAPI.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.issueID = args[0]
			return runIssueRelations(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.category, "category", "", "(required) Issue category: RR/SF/IR/SR/AR/Task/Bug/US/Epic/FE")
	cmd.Flags().StringVar(&o.isSrc, "is-src", "", "Cross-project direction (true|false); omit to let the API decide")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runIssueRelations(cmd *cobra.Command, o *issueRelationsOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}
	if o.category == "" {
		return fmt.Errorf("--category is required (one of RR/SF/IR/SR/AR/Task/Bug/US/Epic/FE)")
	}
	var isSrc *bool
	switch strings.ToLower(strings.TrimSpace(o.isSrc)) {
	case "":
		// omit
	case "true", "1", "yes":
		v := true
		isSrc = &v
	case "false", "0", "no":
		v := false
		isSrc = &v
	default:
		return fmt.Errorf("--is-src must be true or false, got %q", o.isSrc)
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		q := map[string]interface{}{
			"issue_id": o.issueID,
			"category": o.category,
		}
		if isSrc != nil {
			q["is_src"] = *isSrc
		}
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "GET",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/e2e/graphs", projectID),
			"project_id": projectID,
			"query":      q,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListE2EGraphs(context.Background(), projectID, o.issueID, o.category, isSrc)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- issue members -----------------------

type issueMembersOpts struct {
	dryRun bool
}

func newIssueMembersCmd() *cobra.Command {
	o := &issueMembersOpts{}
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List project members (ListProjectUsers)",
		Long: `List all members of the configured project.

Project is taken from config.projectId — set it with
` + "`codearts-cli config set projectId <uuid>`" + ` if unset.

Each returned user includes user_id / user_name / nick_name / role_name;
user_id is the 32-char UUID you need for --assignee on ` + "`issue create`" + `
and for assignee filters on ` + "`issue list`" + `.

API reference: https://support.huaweicloud.com/api-projectman/ListProjectUsers.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueMembers(cmd, o)
		},
	}
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runIssueMembers(cmd *cobra.Command, o *issueMembersOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "GET",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/users", projectID),
			"project_id": projectID,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListProjectUsers(context.Background(), projectID)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- issue list -----------------------

type issueListOpts struct {
	issueType  string
	filterJSON string
	filterFile string
	filterMode string
	pageNo     int
	pageSize   int
	sortField  string
	sortAsc    bool
	dryRun     bool
}

func newIssueListCmd() *cobra.Command {
	o := &issueListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List project work items (ListIpdProjectIssues)",
		Long: `List work items in a project.

--issue-type is required. Multiple types: comma-separated.
Valid types depend on project kind:
  - Systems/Devices: RR,SF,IR,SR,AR,Task,Bug
  - Independent Software: RR,SF,IR,US,Task,Bug
  - Cloud Services: RR,Epic,FE,US,Task,Bug

Filter schema (Huawei IPD API): array of {"<field>": {"values":[...], "operator":"||"}}.
  operator: "||" (OR, default) | "!" (NOT) | "=" | "<>" | "<" | ">"
  example: filter my own bugs →
    --filter '[{"assignee":{"values":["<user_id>"],"operator":"||"}}]'

API reference: https://support.huaweicloud.com/api-projectman/ListIpdProjectIssues.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueList(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.issueType, "issue-type", "", "(required) issue type(s), comma-separated")
	cmd.Flags().StringVar(&o.filterJSON, "filter", "", "JSON array of filter conditions")
	cmd.Flags().StringVar(&o.filterFile, "filter-file", "", "Path to a JSON file containing the filter array")
	cmd.Flags().StringVar(&o.filterMode, "filter-mode", "", "AND_OR (default) or OR_AND")
	cmd.Flags().IntVar(&o.pageNo, "page-no", 0, "Page number (1-based; 0 = API default)")
	cmd.Flags().IntVar(&o.pageSize, "page-size", 0, "Page size (0 = API default)")
	cmd.Flags().StringVar(&o.sortField, "sort-field", "", "Sort by this field (optional)")
	cmd.Flags().BoolVar(&o.sortAsc, "sort-asc", false, "Ascending sort (default desc)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	_ = cmd.MarkFlagRequired("issue-type")
	return cmd
}

func runIssueList(cmd *cobra.Command, o *issueListOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}

	body := &client.ListIssuesRequest{
		FilterMode: o.filterMode,
	}
	rawFilter, err := FirstNonEmpty("--filter", o.filterJSON, "--filter-file", o.filterFile)
	if err != nil {
		return err
	}
	if rawFilter != "" {
		if err := json.Unmarshal([]byte(rawFilter), &body.Filter); err != nil {
			return fmt.Errorf("parse --filter JSON: %w", err)
		}
	}
	if o.pageNo > 0 || o.pageSize > 0 {
		body.Page = &client.PageInfo{PageNo: o.pageNo, PageSize: o.pageSize}
	}
	if o.sortField != "" {
		body.Sort = []client.SortInfo{{Field: o.sortField, Asc: o.sortAsc}}
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "POST",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/query", projectID),
			"project_id": projectID,
			"query":      map[string]string{"issue_type": o.issueType},
			"body":       body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ListIpdProjectIssues(context.Background(), projectID, o.issueType, body)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- issue show -----------------------

type issueShowOpts struct {
	issueID   string
	issueType string
	domainID  string
	dryRun    bool
}

func newIssueShowCmd() *cobra.Command {
	o := &issueShowOpts{}
	cmd := &cobra.Command{
		Use:   "show <issue_id>",
		Short: "Show issue detail (ShowIssueDetail)",
		Long: `Get the full detail of a single work item.

--issue-type is required.
API reference: https://support.huaweicloud.com/api-projectman/ShowIssueDetail.html`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.issueID = args[0]
			return runIssueShow(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.issueType, "issue-type", "", "(required) issue type, e.g. US")
	cmd.Flags().StringVar(&o.domainID, "domain-id", "", "Domain ID (optional)")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	_ = cmd.MarkFlagRequired("issue-type")
	return cmd
}

func runIssueShow(cmd *cobra.Command, o *issueShowOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		q := map[string]string{"issue_type": o.issueType}
		if o.domainID != "" {
			q["domain_id"] = o.domainID
		}
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "GET",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/%s", projectID, o.issueID),
			"project_id": projectID,
			"issue_id":   o.issueID,
			"query":      q,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.ShowIssueDetail(context.Background(), projectID, o.issueID, o.issueType, o.domainID)
	if err != nil {
		return err
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- issue create -----------------------

type issueCreateOpts struct {
	title       string
	description string
	category    string
	assignee    string
	status      string
	priority    string
	bodyJSON    string
	bodyFile    string
	dryRun      bool
}

func newIssueCreateCmd() *cobra.Command {
	o := &issueCreateOpts{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a work item (CreateIpdProjectIssue)",
		Long: `Create a work item.

Required (either via flags or --body/--body-file):
  --title, --description, --category, --assignee

Any additional fields (plan_iteration, workload_man_day, business_domain, ...)
can be passed via --body-file which takes a full JSON object.

API reference: https://support.huaweicloud.com/api-projectman/CreateIpdProjectIssue.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueCreate(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.title, "title", "", "Issue title (max 256 chars)")
	cmd.Flags().StringVar(&o.description, "description", "", "Issue description")
	cmd.Flags().StringVar(&o.category, "category", "", "Issue category (RR/SF/IR/SR/AR/Task/Bug/US/Epic/FE)")
	cmd.Flags().StringVar(&o.assignee, "assignee", "", "Assignee user_id (32-char UUID)")
	cmd.Flags().StringVar(&o.status, "status", "", "Status code (optional): Committed | Analyse | ToBeConfirmed | Plan | Doing | Delivered | Checking")
	cmd.Flags().StringVar(&o.priority, "priority", "", "Priority (optional)")
	cmd.Flags().StringVar(&o.bodyJSON, "body", "", "Full JSON body (overrides flag-based fields)")
	cmd.Flags().StringVar(&o.bodyFile, "body-file", "", "Path to a JSON file for the full body")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runIssueCreate(cmd *cobra.Command, o *issueCreateOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}

	var body interface{}
	rawBody, err := FirstNonEmpty("--body", o.bodyJSON, "--body-file", o.bodyFile)
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
		// Default assignee to the caller's saved user_id when omitted —
		// the most common case for self-assigned issues.
		assignee := o.assignee
		if assignee == "" {
			assignee = cfg.UserID
		}
		if o.title == "" || o.description == "" || o.category == "" || assignee == "" {
			missing := []string{}
			if o.title == "" {
				missing = append(missing, "--title")
			}
			if o.description == "" {
				missing = append(missing, "--description")
			}
			if o.category == "" {
				missing = append(missing, "--category")
			}
			if assignee == "" {
				missing = append(missing, "--assignee (or run `codearts-cli config set userId <uuid>`)")
			}
			return fmt.Errorf("missing required fields: %s (or pass --body / --body-file)", strings.Join(missing, ", "))
		}
		body = &client.CreateIssueRequest{
			Title:       o.title,
			Description: o.description,
			Category:    o.category,
			Assignee:    assignee,
			Status:      o.status,
			Priority:    o.priority,
		}
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "POST",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues", projectID),
			"project_id": projectID,
			"body":       body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.CreateIpdProjectIssue(context.Background(), projectID, body)
	if err != nil {
		return err
	}
	// Extract issue ID from response for a more actionable success message.
	issueID := ExtractStringFromResp(resp, "id")
	if issueID != "" {
		output.Successf(cmd.ErrOrStderr(), "Issue created (id: %s)", issueID)
	} else {
		output.Successf(cmd.ErrOrStderr(), "Issue created")
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ExtractStringFromResp tries to pull a named field out of the standard
// Huawei envelope: {"result": [{"id": "..."}]} or {"result": {"id": "..."}}.
func ExtractStringFromResp(resp map[string]interface{}, key string) string {
	if v, ok := resp[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	result, _ := resp["result"]
	switch r := result.(type) {
	case []interface{}:
		if len(r) > 0 {
			if m, ok := r[0].(map[string]interface{}); ok {
				if s, ok := m[key].(string); ok {
					return s
				}
			}
		}
	case map[string]interface{}:
		if s, ok := r[key].(string); ok {
			return s
		}
	}
	return ""
}

// ----------------------- issue batch-update -----------------------

type issueBatchOpts struct {
	ids          []string
	category     string
	attribute    string
	attributeFile string
	dryRun       bool
}

func newIssueBatchUpdateCmd() *cobra.Command {
	o := &issueBatchOpts{}
	cmd := &cobra.Command{
		Use:   "batch-update",
		Short: "Batch update work items (BatchUpdateIpdIssues)",
		Long: `Apply the same attribute changes to many work items.

At minimum you must pass --id (repeatable) and --category. Additional
attribute fields (status, priority, …) can be supplied via --attribute
(inline JSON) or --attribute-file; category in the file overrides --category.

API reference: https://support.huaweicloud.com/api-projectman/BatchUpdateIpdIssues.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueBatchUpdate(cmd, o)
		},
	}
	cmd.Flags().StringSliceVar(&o.ids, "id", nil, "Issue ID to update (repeatable)")
	cmd.Flags().StringVar(&o.category, "category", "", "Target category (required unless present in --attribute)")
	cmd.Flags().StringVar(&o.attribute, "attribute", "", "JSON object of attributes to set")
	cmd.Flags().StringVar(&o.attributeFile, "attribute-file", "", "Path to a JSON file for the attribute object")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runIssueBatchUpdate(cmd *cobra.Command, o *issueBatchOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}
	if len(o.ids) == 0 {
		return fmt.Errorf("at least one --id is required")
	}
	// Split comma-separated values too so --id a,b,c works the same as
	// --id a --id b --id c.
	flatIDs := make([]string, 0, len(o.ids))
	for _, entry := range o.ids {
		for _, s := range strings.Split(entry, ",") {
			if v := strings.TrimSpace(s); v != "" {
				flatIDs = append(flatIDs, v)
			}
		}
	}
	attr := map[string]interface{}{}
	rawAttr, err := FirstNonEmpty("--attribute", o.attribute, "--attribute-file", o.attributeFile)
	if err != nil {
		return err
	}
	if rawAttr != "" {
		if err := json.Unmarshal([]byte(rawAttr), &attr); err != nil {
			return fmt.Errorf("parse --attribute JSON: %w", err)
		}
	}
	if o.category != "" {
		attr["category"] = o.category
	}
	if _, ok := attr["category"]; !ok || attr["category"] == "" {
		return fmt.Errorf("attribute.category is required (pass --category or include it in --attribute)")
	}

	body := &client.BatchUpdateIssuesRequest{IDs: flatIDs, Attribute: attr}
	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "PUT",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/batch", projectID),
			"project_id": projectID,
			"body":       body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.BatchUpdateIpdIssues(context.Background(), projectID, body)
	if err != nil {
		return err
	}
	output.Successf(cmd.ErrOrStderr(), "Batch update submitted for %d issue(s)", len(flatIDs))
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}

// ----------------------- issue comment -----------------------

func newIssueCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Comment operations on a work item",
	}
	cmd.AddCommand(newIssueCommentAddCmd())
	return cmd
}

type issueCommentAddOpts struct {
	issueCategory   string
	description     string
	descriptionFile string
	bodyJSON        string
	bodyFile        string
	dryRun          bool
}

func newIssueCommentAddCmd() *cobra.Command {
	o := &issueCommentAddOpts{}
	cmd := &cobra.Command{
		Use:   "add <issue_id>",
		Short: "Add a comment to a work item (CreateIssueComment)",
		Long: `Post a comment to an IPD work item.

Required:
  <issue_id>          positional, the 18–19 digit work-item id (returned as
                      "id" by issue list/show — NOT the short "number")
  --issue-category    Task | Bug | US | RR | SF | IR | SR | AR | Epic | FE
  --description       inline HTML body (use --description-file for long text)

The endpoint is undocumented in Huawei's public API reference; the body
shape was reverse-engineered from the CodeArts UI and verified end-to-end.
description is HTML — wrap plain text as <p>…</p>.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runIssueCommentAdd(cmd, args[0], o)
		},
	}
	cmd.Flags().StringVar(&o.issueCategory, "issue-category", "", "Work-item type: Task/Bug/US/RR/SF/IR/SR/AR/Epic/FE")
	cmd.Flags().StringVar(&o.description, "description", "", "Comment HTML body (e.g. \"<p>hello</p>\")")
	cmd.Flags().StringVar(&o.descriptionFile, "description-file", "", "Path to a file whose contents become the description")
	cmd.Flags().StringVar(&o.bodyJSON, "body", "", "Full JSON body (overrides flag-based fields)")
	cmd.Flags().StringVar(&o.bodyFile, "body-file", "", "Path to a JSON file for the full body")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Print the resolved request and exit")
	return cmd
}

func runIssueCommentAdd(cmd *cobra.Command, issueID string, o *issueCommentAddOpts) error {
	cfg, err := core.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	projectID := cfg.ProjectID
	if projectID == "" {
		return fmt.Errorf("no project_id in config — run `codearts-cli config set projectId <uuid>`")
	}

	var body interface{}
	rawBody, err := FirstNonEmpty("--body", o.bodyJSON, "--body-file", o.bodyFile)
	if err != nil {
		return err
	}
	if rawBody != "" {
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(rawBody), &m); err != nil {
			return fmt.Errorf("parse body JSON: %w", err)
		}
		// category defaults to "comment" if caller forgot — it's the only valid value here.
		if _, ok := m["category"]; !ok {
			m["category"] = "comment"
		}
		body = m
	} else {
		desc, err := FirstNonEmpty("--description", o.description, "--description-file", o.descriptionFile)
		if err != nil {
			return err
		}
		if o.issueCategory == "" || desc == "" {
			missing := []string{}
			if o.issueCategory == "" {
				missing = append(missing, "--issue-category")
			}
			if desc == "" {
				missing = append(missing, "--description (or --description-file)")
			}
			return fmt.Errorf("missing required fields: %s (or pass --body / --body-file)", strings.Join(missing, ", "))
		}
		body = &client.CreateIssueCommentRequest{
			Category:      "comment",
			IssueCategory: o.issueCategory,
			Description:   desc,
		}
	}

	if o.dryRun {
		output.DryRunf(cmd.ErrOrStderr(), "request preview (not sent)")
		output.PrintJSON(cmd.OutOrStdout(), map[string]interface{}{
			"method":     "POST",
			"path":       fmt.Sprintf("/v1/ipdprojectservice/projects/%s/issues/%s/comments", projectID, issueID),
			"project_id": projectID,
			"issue_id":   issueID,
			"body":       body,
		})
		return nil
	}
	cli, err := client.New(cfg)
	if err != nil {
		return err
	}
	resp, err := cli.CreateIssueComment(context.Background(), projectID, issueID, body)
	if err != nil {
		return err
	}
	commentID := ExtractStringFromResp(resp, "id")
	if commentID != "" {
		output.Successf(cmd.ErrOrStderr(), "Comment posted (id: %s)", commentID)
	} else {
		output.Successf(cmd.ErrOrStderr(), "Comment posted")
	}
	output.PrintJSON(cmd.OutOrStdout(), resp)
	return nil
}
