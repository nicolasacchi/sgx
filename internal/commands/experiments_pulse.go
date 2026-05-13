package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/nicolasacchi/sgx/internal/client"
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

Control and test groups are auto-resolved when --control/--test are omitted.

Examples:
  sgx experiments pulse my_experiment
  sgx experiments pulse my_experiment --no-cuped --confidence 90
  sgx experiments pulse my_experiment --control ctrl_id --test test_id`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		expID := args[0]
		control := pulseControl
		test := pulseTest

		if control == "" || test == "" {
			control, test, err = resolveGroups(ctx, c, expID)
			if err != nil {
				return err
			}
		}

		params := buildPulseParams(
			pulseNoCuped, pulseConfidence, pulseDate,
			pulseBonferroniVar, pulseBonferroniMetric, pulseBonferroniWeight,
			pulseBHMetric, pulseBHVariant,
			control, test,
		)

		cmdArgs := map[string]any{
			"id":      expID,
			"cuped":   !pulseNoCuped,
			"control": control,
			"test":    test,
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+expID+"/pulse_results", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.pulse", cmdArgs, resp.Data, nil)
	},
}

func buildPulseParams(
	noCuped bool, confidence int, date string,
	bonferroniVar, bonferroniMetric bool, bonferroniWeight float64,
	bhMetric, bhVariant bool,
	control, test string,
) url.Values {
	params := url.Values{}
	if noCuped {
		params.Set("cuped", "false")
	} else {
		params.Set("cuped", "true")
	}
	if confidence > 0 {
		params.Set("confidence", fmt.Sprintf("%d", confidence))
	}
	if date != "" {
		params.Set("date", date)
	}
	if bonferroniVar {
		params.Set("applyBonferroniPerVariant", "true")
	}
	if bonferroniMetric {
		params.Set("applyBonferroniPerMetric", "true")
	}
	if bonferroniWeight > 0 {
		params.Set("bonferroniPrimaryMetricWeight", fmt.Sprintf("%f", bonferroniWeight))
	}
	if bhMetric {
		params.Set("applyBenjaminiHochbergPerMetric", "true")
	}
	if bhVariant {
		params.Set("applyBenjaminiHochbergPerVariant", "true")
	}
	params.Set("control", control)
	params.Set("test", test)
	return params
}

func resolveGroups(ctx context.Context, c *client.Client, expID string) (string, string, error) {
	resp, err := c.Get(ctx, "/console/v1/experiments/"+expID, nil)
	if err != nil {
		return "", "", fmt.Errorf("auto-resolve groups: %w", err)
	}

	var exp struct {
		ControlGroupID string `json:"controlGroupID"`
		Groups         []struct {
			ID   string  `json:"id"`
			Name string  `json:"name"`
			Size float64 `json:"size"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(resp.Data, &exp); err != nil {
		return "", "", fmt.Errorf("parse experiment for group resolution: %w", err)
	}

	if exp.ControlGroupID == "" {
		return "", "", fmt.Errorf("experiment %q has no controlGroupID — specify --control and --test manually", expID)
	}

	var testGroups []struct{ id, name string }
	for _, g := range exp.Groups {
		if g.ID != exp.ControlGroupID {
			testGroups = append(testGroups, struct{ id, name string }{g.ID, g.Name})
		}
	}

	if len(testGroups) == 0 {
		return "", "", fmt.Errorf("experiment %q has no test groups", expID)
	}

	if len(testGroups) == 1 {
		if verboseFlag {
			fmt.Fprintf(os.Stderr, "auto-resolved: control=%s test=%s\n", exp.ControlGroupID, testGroups[0].id)
		}
		return exp.ControlGroupID, testGroups[0].id, nil
	}

	// Multiple test groups: use first, warn on stderr
	names := make([]string, len(testGroups))
	for i, g := range testGroups {
		names[i] = fmt.Sprintf("%s (%s)", g.name, g.id)
	}
	fmt.Fprintf(os.Stderr, "auto-resolved: using first of %d test groups: %s\n", len(testGroups), strings.Join(names, ", "))
	fmt.Fprintf(os.Stderr, "specify --test to choose a different group\n")
	return exp.ControlGroupID, testGroups[0].id, nil
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
