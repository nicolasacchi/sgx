package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	auditAction    string
	auditStartDate string
	auditEndDate   string
	auditSort      string
	auditOrder     string
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View audit logs",
	Long: `View audit trail of configuration changes.

Examples:
  sgx audit
  sgx audit --action experiment_start --start-date 2025-02-01
  sgx audit --start-date 2025-02-01 --end-date 2025-02-25 --order asc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if auditAction != "" {
			params.Set("actionType", auditAction)
		}
		if auditStartDate != "" {
			params.Set("startDate", auditStartDate)
		}
		if auditEndDate != "" {
			params.Set("endDate", auditEndDate)
		}
		if auditSort != "" {
			params.Set("sortKey", auditSort)
		}
		if auditOrder != "" {
			params.Set("sortOrder", auditOrder)
		}
		if limitFlag > 0 {
			params.Set("limit", fmt.Sprintf("%d", limitFlag))
		}
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		cmdArgs := map[string]any{}
		if auditAction != "" {
			cmdArgs["action"] = auditAction
		}
		if auditStartDate != "" {
			cmdArgs["start_date"] = auditStartDate
		}

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/audit_logs", params)
			if err != nil {
				return err
			}
			return output.PrintSuccess(getFormat(), "audit", cmdArgs, resp.Data, resp.Pagination)
		}

		data, pagination, err := c.GetAll(ctx, "/console/v1/audit_logs", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "audit", cmdArgs, mergeRawMessages(data), pagination)
	},
}

func init() {
	auditCmd.Flags().StringVar(&auditAction, "action", "", "Filter by action type (e.g. experiment_start, gate_create)")
	auditCmd.Flags().StringVar(&auditStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	auditCmd.Flags().StringVar(&auditEndDate, "end-date", "", "End date (YYYY-MM-DD)")
	auditCmd.Flags().StringVar(&auditSort, "sort", "", "Sort key (default: date)")
	auditCmd.Flags().StringVar(&auditOrder, "order", "", "Sort order: asc or desc (default: desc)")

	rootCmd.AddCommand(auditCmd)
}
