package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var experimentsExposuresCmd = &cobra.Command{
	Use:   "exposures <id>",
	Short: "Get cumulative exposure counts",
	Long: `Get cumulative exposure data for an experiment over time.

Examples:
  sgx experiments exposures my_experiment`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+args[0]+"/cumulative_exposures", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.exposures", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var experimentsVersionsCmd = &cobra.Command{
	Use:   "versions <id>",
	Short: "Get experiment version history",
	Long: `Get version history for an experiment.

Examples:
  sgx experiments versions my_experiment`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if limitFlag > 0 {
			params.Set("limit", fmt.Sprintf("%d", limitFlag))
		}
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+args[0]+"/versions", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.versions", map[string]any{"id": args[0]}, resp.Data, resp.Pagination)
	},
}

var experimentsOverridesCmd = &cobra.Command{
	Use:   "overrides <id>",
	Short: "Get experiment overrides",
	Long: `Get override rules for an experiment.

Examples:
  sgx experiments overrides my_experiment`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+args[0]+"/overrides", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.overrides", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

func init() {
	experimentsCmd.AddCommand(experimentsExposuresCmd)
	experimentsCmd.AddCommand(experimentsVersionsCmd)
	experimentsCmd.AddCommand(experimentsOverridesCmd)
}
