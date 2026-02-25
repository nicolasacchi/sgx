package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	reportType string
	reportDate string
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Download bulk reports",
	Long: `Download bulk report data. Returns a download URL.

Report types:
  - first_exposures
  - pulse_daily
  - topline_impact_daily

Examples:
  sgx reports --type pulse_daily --date 2025-02-20
  sgx reports --type first_exposures --date 2025-02-20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportType == "" {
			return fmt.Errorf("--type is required (first_exposures, pulse_daily, topline_impact_daily)")
		}
		if reportDate == "" {
			return fmt.Errorf("--date is required (YYYY-MM-DD)")
		}

		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("type", reportType)
		params.Set("date", reportDate)

		resp, err := c.Get(ctx, "/console/v1/reports", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "reports", map[string]any{"type": reportType, "date": reportDate}, resp.Data, nil)
	},
}

func init() {
	reportsCmd.Flags().StringVar(&reportType, "type", "", "Report type (required): first_exposures, pulse_daily, topline_impact_daily")
	reportsCmd.Flags().StringVar(&reportDate, "date", "", "Report date (required): YYYY-MM-DD")

	rootCmd.AddCommand(reportsCmd)
}
