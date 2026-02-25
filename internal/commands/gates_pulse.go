package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	gatePulseNoCuped    bool
	gatePulseConfidence int
)

var gatesPulseCmd = &cobra.Command{
	Use:   "pulse <gate-id> <rule-id>",
	Short: "Get pulse results for a gate rule",
	Long: `Get statistical pulse results for a specific gate rule.

Examples:
  sgx gates pulse my_gate rule_123
  sgx gates pulse my_gate rule_123 --no-cuped --confidence 90`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if gatePulseNoCuped {
			params.Set("cuped", "false")
		} else {
			params.Set("cuped", "true")
		}
		if gatePulseConfidence > 0 {
			params.Set("confidence", fmt.Sprintf("%d", gatePulseConfidence))
		}

		path := fmt.Sprintf("/console/v1/gates/%s/rules/%s/pulse_results", args[0], args[1])
		resp, err := c.Get(ctx, path, params)
		if err != nil {
			return err
		}

		cmdArgs := map[string]any{
			"gate_id": args[0],
			"rule_id": args[1],
			"cuped":   !gatePulseNoCuped,
		}
		return output.PrintSuccess(getFormat(), "gates.pulse", cmdArgs, resp.Data, nil)
	},
}

func init() {
	gatesPulseCmd.Flags().BoolVar(&gatePulseNoCuped, "no-cuped", false, "Disable CUPED variance reduction")
	gatesPulseCmd.Flags().IntVar(&gatePulseConfidence, "confidence", 95, "Confidence interval 0-100")

	gatesCmd.AddCommand(gatesPulseCmd)
}
