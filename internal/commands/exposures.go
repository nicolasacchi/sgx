package commands

import (
	"context"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	exposureExperiments string
	exposureGates       string
	exposureConfigs     string
)

var exposuresCmd = &cobra.Command{
	Use:   "exposures",
	Short: "Get exposure counts",
	Long: `Get exposure counts for experiments, gates, and dynamic configs.
If no filters specified, returns all exposure counts.

Examples:
  sgx exposures
  sgx exposures --experiments exp1,exp2
  sgx exposures --gates gate1,gate2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if exposureExperiments != "" {
			params.Set("experiments", exposureExperiments)
		}
		if exposureGates != "" {
			params.Set("gates", exposureGates)
		}
		if exposureConfigs != "" {
			params.Set("dynamicConfigs", exposureConfigs)
		}

		cmdArgs := map[string]any{}
		if exposureExperiments != "" {
			cmdArgs["experiments"] = exposureExperiments
		}
		if exposureGates != "" {
			cmdArgs["gates"] = exposureGates
		}

		resp, err := c.Get(ctx, "/console/v1/exposure_count", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "exposures", cmdArgs, resp.Data, nil)
	},
}

func init() {
	exposuresCmd.Flags().StringVar(&exposureExperiments, "experiments", "", "Experiment IDs (comma-separated)")
	exposuresCmd.Flags().StringVar(&exposureGates, "gates", "", "Gate IDs (comma-separated)")
	exposuresCmd.Flags().StringVar(&exposureConfigs, "configs", "", "Dynamic config IDs (comma-separated)")

	rootCmd.AddCommand(exposuresCmd)
}
