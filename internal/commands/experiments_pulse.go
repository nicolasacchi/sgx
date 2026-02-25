package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	pulseCuped            bool
	pulseNoCuped          bool
	pulseConfidence       int
	pulseDate             string
	pulseBonferroniVar    bool
	pulseBonferroniMetric bool
	pulseBonferroniWeight float64
	pulseBHMetric         bool
	pulseBHVariant        bool
	pulseControl          string
	pulseTest             string
)

var experimentsPulseCmd = &cobra.Command{
	Use:   "pulse <id>",
	Short: "Get pulse results (statistical analysis)",
	Long: `Get full statistical pulse results for an experiment. This is the most
important command — returns statistical significance, confidence intervals,
and metric deltas.

Examples:
  sgx experiments pulse my_experiment
  sgx experiments pulse my_experiment --no-cuped --confidence 90
  sgx experiments pulse my_experiment --date 2025-02-20 --control ctrl --test test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if pulseNoCuped {
			params.Set("cuped", "false")
		} else {
			params.Set("cuped", "true")
		}
		if pulseConfidence > 0 {
			params.Set("confidence", fmt.Sprintf("%d", pulseConfidence))
		}
		if pulseDate != "" {
			params.Set("date", pulseDate)
		}
		if pulseBonferroniVar {
			params.Set("bonferroniPerVariant", "true")
		}
		if pulseBonferroniMetric {
			params.Set("bonferroniPerMetric", "true")
		}
		if pulseBonferroniWeight > 0 {
			params.Set("bonferroniAlphaWeight", fmt.Sprintf("%f", pulseBonferroniWeight))
		}
		if pulseBHMetric {
			params.Set("bhPerMetric", "true")
		}
		if pulseBHVariant {
			params.Set("bhPerVariant", "true")
		}
		if pulseControl != "" {
			params.Set("control", pulseControl)
		}
		if pulseTest != "" {
			params.Set("test", pulseTest)
		}

		cmdArgs := map[string]any{
			"id":    args[0],
			"cuped": !pulseNoCuped,
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+args[0]+"/pulse_results", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.pulse", cmdArgs, resp.Data, nil)
	},
}

func init() {
	experimentsPulseCmd.Flags().BoolVar(&pulseCuped, "cuped", true, "Apply CUPED variance reduction (default: true)")
	experimentsPulseCmd.Flags().BoolVar(&pulseNoCuped, "no-cuped", false, "Disable CUPED variance reduction")
	experimentsPulseCmd.Flags().IntVar(&pulseConfidence, "confidence", 95, "Confidence interval 0-100")
	experimentsPulseCmd.Flags().StringVar(&pulseDate, "date", "", "Specific date (YYYY-MM-DD)")
	experimentsPulseCmd.Flags().BoolVar(&pulseBonferroniVar, "bonferroni-variant", false, "Apply Bonferroni correction per variant")
	experimentsPulseCmd.Flags().BoolVar(&pulseBonferroniMetric, "bonferroni-metric", false, "Apply Bonferroni correction per metric")
	experimentsPulseCmd.Flags().Float64Var(&pulseBonferroniWeight, "bonferroni-weight", 0, "Alpha allocated to primary metrics")
	experimentsPulseCmd.Flags().BoolVar(&pulseBHMetric, "bh-metric", false, "Apply Benjamini-Hochberg per metric")
	experimentsPulseCmd.Flags().BoolVar(&pulseBHVariant, "bh-variant", false, "Apply Benjamini-Hochberg per variant")
	experimentsPulseCmd.Flags().StringVar(&pulseControl, "control", "", "Control group ID")
	experimentsPulseCmd.Flags().StringVar(&pulseTest, "test", "", "Test group ID")

	experimentsCmd.AddCommand(experimentsPulseCmd)
}
