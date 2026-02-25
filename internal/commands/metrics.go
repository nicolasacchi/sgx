package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	metricsTags       string
	metricsShowHidden bool
	metricsDate       string
	metricsName       string
	metricsType       string
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "List and inspect metrics",
	Long: `List and inspect metric definitions, values, and lineage.

Examples:
  sgx metrics list
  sgx metrics get add_to_cart::event_count
  sgx metrics value add_to_cart::event_count --date 2025-02-20`,
}

var metricsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all metrics",
	Long: `List all metric definitions.

Examples:
  sgx metrics list
  sgx metrics list --show-hidden
  sgx metrics list --tags tag1,tag2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if metricsTags != "" {
			params.Set("tags", metricsTags)
		}
		if metricsShowHidden {
			params.Set("showHiddenMetrics", "true")
		}
		if limitFlag > 0 {
			params.Set("limit", fmt.Sprintf("%d", limitFlag))
		}
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/metrics/list", params)
			if err != nil {
				return err
			}
			return output.PrintSuccess(getFormat(), "metrics.list", nil, resp.Data, resp.Pagination)
		}

		data, pagination, err := c.GetAll(ctx, "/console/v1/metrics/list", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.list", nil, mergeRawMessages(data), pagination)
	},
}

var metricsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get metric definition",
	Long: `Get full definition for a specific metric.
The id format is <metric_name>::<type> e.g. add_to_cart::event_count

Examples:
  sgx metrics get add_to_cart::event_count`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/metrics/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.get", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var metricsValueCmd = &cobra.Command{
	Use:   "value <id>",
	Short: "Get single metric value",
	Long: `Get the value of a specific metric.
The id format is <metric_name>::<type> e.g. add_to_cart::event_count

Examples:
  sgx metrics value add_to_cart::event_count
  sgx metrics value add_to_cart::event_count --date 2025-02-20`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("id", args[0])
		if metricsDate != "" {
			params.Set("date", metricsDate)
		}

		resp, err := c.Get(ctx, "/console/v1/metrics", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.value", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var metricsValuesCmd = &cobra.Command{
	Use:   "values",
	Short: "Get all metric values",
	Long: `Get values for all metrics, optionally filtered.

Examples:
  sgx metrics values
  sgx metrics values --date 2025-02-20
  sgx metrics values --name conversion_rate --type ratio`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if metricsDate != "" {
			params.Set("date", metricsDate)
		}
		if metricsName != "" {
			params.Set("metricName", metricsName)
		}
		if metricsType != "" {
			params.Set("metricType", metricsType)
		}
		if limitFlag > 0 {
			params.Set("limit", fmt.Sprintf("%d", limitFlag))
		}
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/metrics/values", params)
			if err != nil {
				return err
			}
			return output.PrintSuccess(getFormat(), "metrics.values", nil, resp.Data, resp.Pagination)
		}

		data, pagination, err := c.GetAll(ctx, "/console/v1/metrics/values", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.values", nil, mergeRawMessages(data), pagination)
	},
}

var metricsExperimentsCmd = &cobra.Command{
	Use:   "experiments <id>",
	Short: "Get experiments using this metric",
	Long: `Get which experiments use this metric (lineage query).

Examples:
  sgx metrics experiments add_to_cart::event_count`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/metrics/"+args[0]+"/experiments", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.experiments", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var metricsSQLCmd = &cobra.Command{
	Use:   "sql <id>",
	Short: "Get metric SQL definition",
	Long: `Get the SQL definition for a metric.

Examples:
  sgx metrics sql add_to_cart::event_count`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/metrics/"+args[0]+"/sql", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.sql", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var metricsSourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "List metric sources",
	Long: `List all metric sources.

Examples:
  sgx metrics sources`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/metrics/metric_source/list", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "metrics.sources", nil, resp.Data, nil)
	},
}

func init() {
	metricsListCmd.Flags().StringVar(&metricsTags, "tags", "", "Filter by tags (comma-separated)")
	metricsListCmd.Flags().BoolVar(&metricsShowHidden, "show-hidden", false, "Include hidden metrics")
	metricsValueCmd.Flags().StringVar(&metricsDate, "date", "", "Specific date (YYYY-MM-DD)")
	metricsValuesCmd.Flags().StringVar(&metricsDate, "date", "", "Specific date (YYYY-MM-DD)")
	metricsValuesCmd.Flags().StringVar(&metricsName, "name", "", "Filter by metric name")
	metricsValuesCmd.Flags().StringVar(&metricsType, "type", "", "Filter by metric type")

	metricsCmd.AddCommand(metricsListCmd)
	metricsCmd.AddCommand(metricsGetCmd)
	metricsCmd.AddCommand(metricsValueCmd)
	metricsCmd.AddCommand(metricsValuesCmd)
	metricsCmd.AddCommand(metricsExperimentsCmd)
	metricsCmd.AddCommand(metricsSQLCmd)
	metricsCmd.AddCommand(metricsSourcesCmd)
	rootCmd.AddCommand(metricsCmd)
}
