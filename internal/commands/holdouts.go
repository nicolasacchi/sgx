package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	holdoutPulseNoCuped    bool
	holdoutPulseConfidence int
)

var holdoutsCmd = &cobra.Command{
	Use:   "holdouts",
	Short: "List and inspect holdouts",
	Long: `List and inspect experiment holdout groups and their pulse results.

Examples:
  sgx holdouts list
  sgx holdouts get my_holdout
  sgx holdouts pulse my_holdout`,
}

var holdoutsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all holdouts",
	Long: `List all holdout groups.

Examples:
  sgx holdouts list`,
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

		resp, err := c.Get(ctx, "/console/v1/holdouts", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "holdouts.list", nil, resp.Data, resp.Pagination)
	},
}

var holdoutsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get holdout details",
	Long: `Get details for a specific holdout group.

Examples:
  sgx holdouts get my_holdout`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/holdouts/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "holdouts.get", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var holdoutsPulseCmd = &cobra.Command{
	Use:   "pulse <id>",
	Short: "Get holdout pulse results",
	Long: `Get statistical pulse results for a holdout group.

Examples:
  sgx holdouts pulse my_holdout
  sgx holdouts pulse my_holdout --no-cuped`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if holdoutPulseNoCuped {
			params.Set("cuped", "false")
		} else {
			params.Set("cuped", "true")
		}
		if holdoutPulseConfidence > 0 {
			params.Set("confidence", fmt.Sprintf("%d", holdoutPulseConfidence))
		}

		resp, err := c.Get(ctx, "/console/v1/holdouts/"+args[0]+"/pulse_results", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "holdouts.pulse", map[string]any{"id": args[0], "cuped": !holdoutPulseNoCuped}, resp.Data, nil)
	},
}

func init() {
	holdoutsPulseCmd.Flags().BoolVar(&holdoutPulseNoCuped, "no-cuped", false, "Disable CUPED variance reduction")
	holdoutsPulseCmd.Flags().IntVar(&holdoutPulseConfidence, "confidence", 95, "Confidence interval 0-100")

	holdoutsCmd.AddCommand(holdoutsListCmd)
	holdoutsCmd.AddCommand(holdoutsGetCmd)
	holdoutsCmd.AddCommand(holdoutsPulseCmd)
	rootCmd.AddCommand(holdoutsCmd)
}
