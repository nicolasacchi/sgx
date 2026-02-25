package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	expStatus       string
	expTags         string
	expTeam         string
	expStale        bool
	expType         string
	expCreatedAfter string
	expCreatedBefore string
	expCreator      string
)

var experimentsCmd = &cobra.Command{
	Use:   "experiments",
	Short: "List and inspect experiments",
	Long: `List and inspect experiments, pulse results, exposures, and overrides.

Examples:
  sgx experiments list --status active
  sgx experiments get my_experiment
  sgx experiments pulse my_experiment --confidence 95`,
}

var experimentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all experiments",
	Long: `List all experiments with optional filters.

Examples:
  sgx experiments list
  sgx experiments list --status active
  sgx experiments list --tags tag1,tag2 --creator john`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if expStatus != "" {
			params.Set("status", expStatus)
		}
		if expTags != "" {
			params.Set("tags", expTags)
		}
		if expTeam != "" {
			params.Set("teamID", expTeam)
		}
		if expStale {
			params.Set("stale", "true")
		}
		if expType != "" {
			params.Set("experimentType", expType)
		}
		if expCreatedAfter != "" {
			params.Set("createdStartDate", expCreatedAfter)
		}
		if expCreatedBefore != "" {
			params.Set("createdEndDate", expCreatedBefore)
		}
		if expCreator != "" {
			params.Set("creatorName", expCreator)
		}
		if limitFlag > 0 {
			params.Set("limit", fmt.Sprintf("%d", limitFlag))
		}
		if pageFlag > 0 {
			params.Set("page", fmt.Sprintf("%d", pageFlag))
		}

		cmdArgs := map[string]any{}
		if expStatus != "" {
			cmdArgs["status"] = expStatus
		}

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/experiments", params)
			if err != nil {
				return err
			}
			return output.PrintSuccess(getFormat(), "experiments.list", cmdArgs, resp.Data, resp.Pagination)
		}

		data, pagination, err := c.GetAll(ctx, "/console/v1/experiments", params)
		if err != nil {
			return err
		}
		merged := mergeRawMessages(data)
		return output.PrintSuccess(getFormat(), "experiments.list", cmdArgs, merged, pagination)
	},
}

var experimentsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get experiment details",
	Long: `Get full details for a specific experiment.

Examples:
  sgx experiments get my_experiment`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.get", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var experimentsContextCmd = &cobra.Command{
	Use:   "context <id>",
	Short: "Get experiment context",
	Long: `Get context information for an experiment.

Examples:
  sgx experiments context my_experiment`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/experiments/"+args[0]+"/context", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.context", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

func init() {
	experimentsListCmd.Flags().StringVar(&expStatus, "status", "", "Filter by status: setup, active, decision_made, abandoned")
	experimentsListCmd.Flags().StringVar(&expTags, "tags", "", "Filter by tag IDs (comma-separated)")
	experimentsListCmd.Flags().StringVar(&expTeam, "team", "", "Filter by team ID")
	experimentsListCmd.Flags().BoolVar(&expStale, "stale", false, "Only stale experiments")
	experimentsListCmd.Flags().StringVar(&expType, "type", "", "Filter by experiment type")
	experimentsListCmd.Flags().StringVar(&expCreatedAfter, "created-after", "", "Created after date (YYYY-MM-DD)")
	experimentsListCmd.Flags().StringVar(&expCreatedBefore, "created-before", "", "Created before date (YYYY-MM-DD)")
	experimentsListCmd.Flags().StringVar(&expCreator, "creator", "", "Filter by creator name")

	experimentsCmd.AddCommand(experimentsListCmd)
	experimentsCmd.AddCommand(experimentsGetCmd)
	experimentsCmd.AddCommand(experimentsContextCmd)
	rootCmd.AddCommand(experimentsCmd)
}
