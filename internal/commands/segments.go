package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var segmentsCmd = &cobra.Command{
	Use:   "segments",
	Short: "List and inspect segments",
	Long: `List and inspect user segments.

Examples:
  sgx segments list
  sgx segments get my_segment`,
}

var segmentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all segments",
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

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/segments", params)
			if err != nil {
				return err
			}
			return output.PrintSuccess(getFormat(), "segments.list", nil, resp.Data, resp.Pagination)
		}

		data, pagination, err := c.GetAll(ctx, "/console/v1/segments", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "segments.list", nil, mergeRawMessages(data), pagination)
	},
}

var segmentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get segment details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/segments/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "segments.get", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

func init() {
	segmentsCmd.AddCommand(segmentsListCmd)
	segmentsCmd.AddCommand(segmentsGetCmd)
	rootCmd.AddCommand(segmentsCmd)
}
