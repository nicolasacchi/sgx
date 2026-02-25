package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nicolasacchi/sgx/internal/config"
	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	configAddKey     string
	configAddBaseURL string
	configAddFormat  string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage project configurations",
	Long: `Add, remove, list, and switch between Statsig project configurations.

Projects are stored in ~/.config/sgx/config.json.

Examples:
  sgx config add production --api-key console-abc123
  sgx config list
  sgx config use production
  sgx config current`,
}

var configAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add or update a project configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if configAddKey == "" {
			return fmt.Errorf("--api-key is required")
		}
		if err := config.AddProject(name, configAddKey, configAddBaseURL, configAddFormat); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Project %q added.\n", name)
		result := map[string]string{"project": name, "status": "added"}
		data, _ := json.Marshal(result)
		return output.PrintSuccess(getFormat(), "config.add", map[string]any{"name": name}, json.RawMessage(data), nil)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a project configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.RemoveProject(name); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Project %q removed.\n", name)
		result := map[string]string{"project": name, "status": "removed"}
		data, _ := json.Marshal(result)
		return output.PrintSuccess(getFormat(), "config.remove", map[string]any{"name": name}, json.RawMessage(data), nil)
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			result := []any{}
			data, _ := json.Marshal(result)
			return output.PrintSuccess(getFormat(), "config.list", nil, json.RawMessage(data), nil)
		}

		var rows []map[string]any
		if cfg.Projects != nil {
			for name, p := range cfg.Projects {
				rows = append(rows, map[string]any{
					"name":     name,
					"api_key":  config.MaskKey(p.APIKey),
					"base_url": p.BaseURL,
					"format":   p.Format,
					"default":  name == cfg.DefaultProject,
				})
			}
		} else if cfg.APIKey != "" {
			rows = append(rows, map[string]any{
				"name":     "(legacy)",
				"api_key":  config.MaskKey(cfg.APIKey),
				"base_url": cfg.BaseURL,
				"format":   cfg.Format,
				"default":  true,
			})
		}

		if rows == nil {
			rows = []map[string]any{}
		}
		data, _ := json.Marshal(rows)
		return output.PrintSuccess(getFormat(), "config.list", nil, json.RawMessage(data), nil)
	},
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the default project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.SetDefaultProject(name); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Default project set to %q.\n", name)
		result := map[string]string{"project": name, "status": "default"}
		data, _ := json.Marshal(result)
		return output.PrintSuccess(getFormat(), "config.use", map[string]any{"name": name}, json.RawMessage(data), nil)
	},
}

var configCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ListProjects()
		if err != nil {
			return fmt.Errorf("no config file found")
		}

		var name string
		var source string
		if projectFlag != "" {
			name = projectFlag
			source = "flag"
		} else if cfg.DefaultProject != "" {
			name = cfg.DefaultProject
			source = "default"
		} else if cfg.APIKey != "" {
			name = "(legacy)"
			source = "legacy"
		} else {
			return fmt.Errorf("no project configured")
		}

		result := map[string]string{"project": name, "source": source}
		data, _ := json.Marshal(result)
		return output.PrintSuccess(getFormat(), "config.current", nil, json.RawMessage(data), nil)
	},
}

func init() {
	configAddCmd.Flags().StringVar(&configAddKey, "api-key", "", "Statsig Console API key (required)")
	configAddCmd.Flags().StringVar(&configAddBaseURL, "base-url", "", "API base URL for this project")
	configAddCmd.Flags().StringVar(&configAddFormat, "format", "", "Default output format for this project")

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configCurrentCmd)
	rootCmd.AddCommand(configCmd)
}
