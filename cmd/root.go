// Package cmd wires the cobra command tree for codearts-cli.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

const rootLong = `codearts-cli — Huawei Cloud CodeArts CLI.

USAGE:
    codearts-cli <command> [subcommand] [options]

QUICK START:
    # 1. Configure AK/SK (interactive)
    codearts-cli config init

    # 2. Show saved config (secrets masked)
    codearts-cli config show

    # 3. Trigger / stop a pipeline run
    codearts-cli pipeline run  <pipeline_id>
    codearts-cli pipeline stop <pipeline_id> <pipeline_run_id>

    # 4. Work items (ProjectMan IPD)
    codearts-cli issue list   --issue-type US
    codearts-cli issue show   <issue_id> --issue-type US
    codearts-cli issue create --title "..." --description "..." --category US --assignee <user_id>
    codearts-cli issue batch-update --id a,b,c --category US
    codearts-cli issue relations <issue_id> --category US
    codearts-cli issue members

    # 5. 代码托管 (CodeArts Repo)
    codearts-cli repo mr create  <repo_id> --title "..." --source feat/x --target main
    codearts-cli repo mr comment <repo_id> <mr_iid> --body "LGTM"
    codearts-cli repo member list <repo_id>

    # 6. 编译构建 (CodeArts Build)
    codearts-cli build list --project-id <proj>
    codearts-cli build run  <job_id> --branch main --build-type branch
    codearts-cli build stop <job_id> <build_no>

DEFAULTS:
    project_id  cd130bd8357b4e7ab293a7979d1c8711
    gateway     http://10.250.63.100:8099

Run 'codearts-cli <command> --help' for details.`

// Execute is the main entrypoint; returns the process exit code.
func Execute() int {
	root := &cobra.Command{
		Use:           "codearts-cli",
		Short:         "Huawei Cloud CodeArts CLI",
		Long:          rootLong,
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(newConfigCmd())
	root.AddCommand(newPipelineCmd())
	root.AddCommand(newIssueCmd())
	root.AddCommand(newRepoCmd())
	root.AddCommand(newBuildCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
