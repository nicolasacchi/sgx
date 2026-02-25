package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	expStatus        string
	expTags          string
	expTeam          string
	expStale         bool
	expType          string
	expCreatedAfter  string
	expCreatedBefore string
	expCreator       string
	expOwner         string
	expSince         string
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

		var finalData json.RawMessage
		var pagination any

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/experiments", params)
			if err != nil {
				return err
			}
			finalData = resp.Data
			pagination = resp.Pagination
		} else {
			data, pg, err := c.GetAll(ctx, "/console/v1/experiments", params)
			if err != nil {
				return err
			}
			finalData = mergeRawMessages(data)
			pagination = pg
		}

		finalData, err = filterExperiments(finalData, expOwner, expSince)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "experiments.list", cmdArgs, finalData, pagination)
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

func filterExperiments(data json.RawMessage, owner, since string) (json.RawMessage, error) {
	if owner == "" && since == "" {
		return data, nil
	}

	var experiments []map[string]any
	if err := json.Unmarshal(data, &experiments); err != nil {
		return data, nil
	}

	var sinceTime time.Time
	if since != "" {
		t, err := time.Parse("2006-01-02", since)
		if err != nil {
			return nil, fmt.Errorf("invalid --since date: %w", err)
		}
		sinceTime = t
	}

	filtered := make([]map[string]any, 0)
	for _, exp := range experiments {
		if owner != "" {
			creator, _ := exp["creatorName"].(string)
			creatorEmail, _ := exp["creatorEmail"].(string)
			if !strings.Contains(strings.ToLower(creator), strings.ToLower(owner)) &&
				!strings.Contains(strings.ToLower(creatorEmail), strings.ToLower(owner)) {
				continue
			}
		}
		if since != "" {
			var latest time.Time
			for _, field := range []string{"lastModifiedTime", "createdTime"} {
				if dateStr, ok := exp[field].(string); ok {
					if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
						if t.After(latest) {
							latest = t
						}
					}
				}
			}
			if !latest.IsZero() && latest.Before(sinceTime) {
				continue
			}
		}
		filtered = append(filtered, exp)
	}

	result, _ := json.Marshal(filtered)
	return result, nil
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
	experimentsListCmd.Flags().StringVar(&expOwner, "owner", "", "Filter by owner/creator (client-side, substring match)")
	experimentsListCmd.Flags().StringVar(&expSince, "since", "", "Only experiments modified after date (YYYY-MM-DD, client-side)")

	experimentsCmd.AddCommand(experimentsListCmd)
	experimentsCmd.AddCommand(experimentsGetCmd)
	experimentsCmd.AddCommand(experimentsContextCmd)
	rootCmd.AddCommand(experimentsCmd)
}
