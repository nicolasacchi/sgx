package commands

import (
	"github.com/nicolasacchi/sgx/internal/client"
	"github.com/nicolasacchi/sgx/internal/config"
	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	version     = "dev"
	apiKeyFlag  string
	formatFlag  string
	baseURLFlag string
	projectFlag string
	verboseFlag bool
	noPaginate  bool
	pageFlag    int
	limitFlag   int
)

var rootCmd = &cobra.Command{
	Use:   "sgx",
	Short: "Statsig Explorer — readonly CLI for experiments, metrics, and stats",
	Long: `sgx (statsig explorer) is a readonly CLI that gives a complete view of a
Statsig project's experiment and metrics state. Every command emits clean
structured JSON to stdout.

Usage examples:
  sgx experiments list --status active
  sgx experiments pulse my_experiment
  sgx gates list
  sgx metrics list
  sgx overview`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func getClient(cmd *cobra.Command) (*client.Client, error) {
	apiKey, err := config.LoadAPIKey(apiKeyFlag, projectFlag)
	if err != nil {
		return nil, err
	}
	baseURL := config.LoadBaseURL(baseURLFlag, projectFlag)
	return client.New(apiKey, baseURL, verboseFlag), nil
}

func getFormat() output.Format {
	f := config.LoadFormat(formatFlag, projectFlag)
	switch output.Format(f) {
	case output.FormatTable:
		return output.FormatTable
	case output.FormatCompact:
		return output.FormatCompact
	default:
		return output.FormatJSON
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Statsig Console API key (overrides STATSIG_API_KEY env var)")
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Use a named project from ~/.config/sgx/config.json")
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "", "Output format: json (default), table, compact")
	rootCmd.PersistentFlags().StringVar(&baseURLFlag, "base-url", "", "API base URL (default: https://statsigapi.net)")
	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Print request details to stderr")
	rootCmd.PersistentFlags().BoolVar(&noPaginate, "no-paginate", false, "Don't auto-paginate, return first page only")
	rootCmd.PersistentFlags().IntVar(&pageFlag, "page", 0, "Specific page number (disables auto-pagination)")
	rootCmd.PersistentFlags().IntVar(&limitFlag, "limit", 100, "Results per page (default: 100, max: 100)")
}
