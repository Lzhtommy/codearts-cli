package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/autelrobotics/codearts-cli/internal/core"
	"github.com/autelrobotics/codearts-cli/internal/output"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage credentials and tenant defaults",
	}
	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigPathCmd())
	cmd.AddCommand(newConfigSetCmd())
	return cmd
}

// ------------------------------ config set ------------------------------

// newConfigSetCmd implements `codearts-cli config set <key> <value>`:
// updates a single field in-place without walking the full init flow.
// This is the right tool when you just want to attach a user_id to an
// already-working AK/SK config.
func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Update a single config field (ak | sk | projectId | region | userId | endpoint)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.ToLower(args[0])
			val := args[1]
			cfg, err := core.Load()
			if err != nil {
				return err
			}
			switch key {
			case "ak":
				cfg.AK = val
			case "sk":
				cfg.SK = val
			case "projectid", "project_id", "project-id":
				cfg.ProjectID = val
			case "region":
				cfg.Region = val
			case "userid", "user_id", "user-id":
				cfg.UserID = val
			case "endpoint":
				cfg.Endpoint = val
			default:
				return fmt.Errorf("unknown key %q; valid keys: ak, sk, projectId, region, userId, endpoint", args[0])
			}
			if err := core.Save(cfg); err != nil {
				return err
			}
			output.Successf(cmd.ErrOrStderr(), "Updated %s", key)
			output.PrintJSON(cmd.OutOrStdout(), core.Redacted(cfg))
			return nil
		},
	}
}

// ------------------------------ config init ------------------------------

type configInitOpts struct {
	ak        string
	sk        string
	skStdin   bool
	projectID string
	region    string
	userID    string
	endpoint  string
	yes       bool
}

func newConfigInitCmd() *cobra.Command {
	o := &configInitOpts{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize AK/SK credentials and tenant defaults",
		Long: `Initialize AK/SK credentials for Huawei Cloud CodeArts.

By default this is interactive — it prompts for AK, SK (hidden), project_id,
and region. You can skip prompts with flags, e.g. for CI:

    echo "$HW_SK" | codearts-cli config init \
        --ak "$HW_AK" --sk-stdin --yes

The config is stored at ~/.codearts-cli/config.json with mode 0600.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(cmd, o)
		},
	}
	cmd.Flags().StringVar(&o.ak, "ak", "", "Access Key ID (non-interactive)")
	cmd.Flags().StringVar(&o.sk, "sk", "", "Secret Access Key (INSECURE — prefer --sk-stdin)")
	cmd.Flags().BoolVar(&o.skStdin, "sk-stdin", false, "Read SK from stdin (avoids process-list exposure)")
	cmd.Flags().StringVar(&o.projectID, "project-id", core.DefaultProjectID, "Huawei Cloud project ID")
	cmd.Flags().StringVar(&o.region, "region", core.DefaultRegion, "Region, e.g. cn-south-1")
	cmd.Flags().StringVar(&o.userID, "user-id", "", "IAM user_id (32-char UUID); used as default assignee/author for write APIs")
	cmd.Flags().StringVar(&o.endpoint, "endpoint", "", "Optional endpoint override (e.g. https://cloudpipeline-ext.cn-south-1.myhuaweicloud.com)")
	cmd.Flags().BoolVarP(&o.yes, "yes", "y", false, "Skip overwrite confirmation if a config already exists")
	return cmd
}

func runConfigInit(cmd *cobra.Command, o *configInitOpts) error {
	// Load existing config (if any) so re-running init preserves values for
	// fields the user leaves blank.
	existing, _ := core.Load()

	ak := o.ak
	sk := o.sk
	projectID := o.projectID
	region := o.region
	endpoint := o.endpoint

	// Read SK from stdin if requested. This path is the recommended one for
	// scripts / CI because the secret never appears in the process list.
	if o.skStdin {
		b, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return fmt.Errorf("read sk from stdin: %w", err)
		}
		sk = strings.TrimSpace(string(b))
		if sk == "" {
			return errors.New("stdin was empty, expected SK")
		}
	}

	// Interactive fill for anything still missing.
	reader := bufio.NewReader(cmd.InOrStdin())
	interactive := term.IsTerminal(int(os.Stdin.Fd()))

	if ak == "" {
		if !interactive {
			return errors.New("--ak is required when stdin is not a terminal")
		}
		def := existing.AK
		v, err := promptLine(cmd, reader, "Access Key (AK)", def, false)
		if err != nil {
			return err
		}
		ak = v
	}
	if sk == "" {
		if !interactive {
			return errors.New("--sk-stdin or --sk is required when stdin is not a terminal")
		}
		hasDefault := existing.SK != ""
		v, err := promptSecret(cmd, "Secret Access Key (SK)", hasDefault)
		if err != nil {
			return err
		}
		if v == "" && hasDefault {
			sk = existing.SK
		} else {
			sk = v
		}
	}
	if !cmd.Flags().Changed("project-id") && interactive {
		def := projectID
		if existing.ProjectID != "" {
			def = existing.ProjectID
		}
		v, err := promptLine(cmd, reader, "Project ID", def, true)
		if err != nil {
			return err
		}
		projectID = v
	}
	if !cmd.Flags().Changed("region") && interactive {
		def := region
		if existing.Region != "" {
			def = existing.Region
		}
		v, err := promptLine(cmd, reader, "Region", def, true)
		if err != nil {
			return err
		}
		region = v
	}
	userID := o.userID
	if !cmd.Flags().Changed("user-id") && interactive {
		def := userID
		if existing.UserID != "" {
			def = existing.UserID
		}
		v, err := promptLine(cmd, reader, "User ID (IAM user_id, 32-char UUID; blank to skip)", def, true)
		if err != nil {
			return err
		}
		userID = v
	}
	if !cmd.Flags().Changed("endpoint") && interactive {
		def := endpoint
		if existing.Endpoint != "" {
			def = existing.Endpoint
		}
		v, err := promptLine(cmd, reader, "Endpoint (optional, leave blank for default)", def, true)
		if err != nil {
			return err
		}
		endpoint = v
	}

	cfg := &core.Config{
		AK:        strings.TrimSpace(ak),
		SK:        strings.TrimSpace(sk),
		ProjectID: strings.TrimSpace(projectID),
		Region:    strings.TrimSpace(region),
		UserID:    strings.TrimSpace(userID),
		Endpoint:  strings.TrimSpace(endpoint),
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Overwrite confirmation if an AK is already present and --yes was not
	// passed. We don't guard on first-run (empty existing) to keep the happy
	// path frictionless.
	if existing.AK != "" && !o.yes && interactive {
		p, _ := core.Path()
		fmt.Fprintf(cmd.ErrOrStderr(), "An existing config is at %s. Overwrite? [y/N] ", p)
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("confirm overwrite: %w", err)
		}
		line = strings.ToLower(strings.TrimSpace(line))
		if line != "y" && line != "yes" {
			return errors.New("aborted")
		}
	}

	if err := core.Save(cfg); err != nil {
		return err
	}
	p, _ := core.Path()
	output.Successf(cmd.ErrOrStderr(), "Configuration saved to %s", p)
	output.PrintJSON(cmd.OutOrStdout(), core.Redacted(cfg))
	return nil
}

// ------------------------------ config show ------------------------------

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print the current configuration (secrets masked)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := core.Load()
			if err != nil {
				return err
			}
			if cfg.AK == "" {
				return errors.New("no configuration found — run `codearts-cli config init`")
			}
			output.PrintJSON(cmd.OutOrStdout(), core.Redacted(cfg))
			return nil
		},
	}
}

// ------------------------------ config path ------------------------------

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := core.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
}

// ------------------------------ prompt helpers ------------------------------

// promptLine reads a single line with an optional default shown in brackets.
// If allowBlank is false, empty input that also lacks a default re-prompts.
func promptLine(cmd *cobra.Command, r *bufio.Reader, label, def string, allowBlank bool) (string, error) {
	for {
		if def != "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s [%s]: ", label, def)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s: ", label)
		}
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("read %s: %w", label, err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			if def != "" {
				return def, nil
			}
			if allowBlank {
				return "", nil
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%s cannot be empty\n", label)
			continue
		}
		return line, nil
	}
}

// promptSecret reads a secret without echo. If hasDefault is true, a blank
// entry signals "keep existing" and returns "".
func promptSecret(cmd *cobra.Command, label string, hasDefault bool) (string, error) {
	if hasDefault {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s [****, blank to keep]: ", label)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s: ", label)
	}
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", fmt.Errorf("read %s: %w", label, err)
	}
	return strings.TrimSpace(string(b)), nil
}
