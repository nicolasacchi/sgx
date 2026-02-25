package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	gateType            string
	gateTypeReason      string
	gateTags            string
	gateIDType          string
	gateIncludeArchived bool
)

var gatesCmd = &cobra.Command{
	Use:   "gates",
	Short: "List and inspect feature gates",
	Long: `List and inspect feature gates, rules, and pulse results.

Examples:
  sgx gates list
  sgx gates get my_gate
  sgx gates rules my_gate`,
}

var gatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all gates",
	Long: `List all feature gates with optional filters.

Examples:
  sgx gates list
  sgx gates list --type STALE
  sgx gates list --include-archived`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if gateType != "" {
			params.Set("type", gateType)
		}
		if gateTypeReason != "" {
			params.Set("typeReason", gateTypeReason)
		}
		if gateTags != "" {
			params.Set("tags", gateTags)
		}
		if gateIDType != "" {
			params.Set("idType", gateIDType)
		}
		if gateIncludeArchived {
			params.Set("includeArchived", "true")
		}
		if limitFlag > 0 {
			params.Set("limit", fmt.Sprintf("%d", limitFlag))
		}
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		cmdArgs := map[string]any{}
		if gateType != "" {
			cmdArgs["type"] = gateType
		}

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/gates", params)
			if err != nil {
				return err
			}
			return output.PrintSuccess(getFormat(), "gates.list", cmdArgs, resp.Data, resp.Pagination)
		}

		data, pagination, err := c.GetAll(ctx, "/console/v1/gates", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "gates.list", cmdArgs, mergeRawMessages(data), pagination)
	},
}

var gatesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get gate details",
	Long: `Get full details for a specific gate.

Examples:
  sgx gates get my_gate`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/gates/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "gates.get", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var gatesRulesCmd = &cobra.Command{
	Use:   "rules <id>",
	Short: "Get gate rules",
	Long: `Get all rules for a specific gate.

Examples:
  sgx gates rules my_gate`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/gates/"+args[0]+"/rules", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "gates.rules", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var gatesChecksCmd = &cobra.Command{
	Use:   "checks <id>",
	Short: "Get gate check counts",
	Long: `Get check counts for a specific gate.

Examples:
  sgx gates checks my_gate`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/gates/"+args[0]+"/checks", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "gates.checks", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

func init() {
	gatesListCmd.Flags().StringVar(&gateType, "type", "", "Filter by type (STALE, PERMANENT, etc.)")
	gatesListCmd.Flags().StringVar(&gateTypeReason, "type-reason", "", "Filter by type reason")
	gatesListCmd.Flags().StringVar(&gateTags, "tags", "", "Filter by tags (comma-separated)")
	gatesListCmd.Flags().StringVar(&gateIDType, "id-type", "", "Filter by ID type")
	gatesListCmd.Flags().BoolVar(&gateIncludeArchived, "include-archived", false, "Include archived gates")

	gatesCmd.AddCommand(gatesListCmd)
	gatesCmd.AddCommand(gatesGetCmd)
	gatesCmd.AddCommand(gatesRulesCmd)
	gatesCmd.AddCommand(gatesChecksCmd)
	rootCmd.AddCommand(gatesCmd)
}
