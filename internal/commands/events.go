package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List and inspect events",
	Long: `List and inspect logged events and their derived metrics.

Examples:
  sgx events list
  sgx events get purchase
  sgx events metrics purchase`,
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List logged events",
	Long: `List all logged events. Auto-pagination is disabled by default for events.

Examples:
  sgx events list
  sgx events list --limit 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		limit := limitFlag
		if limit == 100 {
			limit = 50 // Default to 50 for events
		}
		params.Set("limit", fmt.Sprintf("%d", limit))
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		// Events: no auto-pagination by default
		resp, err := c.Get(ctx, "/console/v1/events", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "events.list", nil, resp.Data, resp.Pagination)
	},
}

var eventsGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get event details",
	Long: `Get details for a specific event by name.

Examples:
  sgx events get purchase`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/events/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "events.get", map[string]any{"name": args[0]}, resp.Data, nil)
	},
}

var eventsMetricsCmd = &cobra.Command{
	Use:   "metrics <name>",
	Short: "Get metrics derived from this event",
	Long: `Get which metrics are derived from a specific event.

Examples:
  sgx events metrics purchase`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/events/"+args[0]+"/metrics", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "events.metrics", map[string]any{"name": args[0]}, resp.Data, nil)
	},
}

func init() {
	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsGetCmd)
	eventsCmd.AddCommand(eventsMetricsCmd)
	rootCmd.AddCommand(eventsCmd)
}
