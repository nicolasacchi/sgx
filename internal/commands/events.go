package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	eventsSince string
	eventsUntil string
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
		finalData, err := filterEventsByTime(resp.Data, eventsSince, eventsUntil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "events.list", nil, finalData, resp.Pagination)
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

type catalogEntry struct {
	Name       string `json:"name"`
	TotalCount int64  `json:"total_count"`
	Entries    int    `json:"entries"`
}

var eventsCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Deduplicated event catalog",
	Long: `Fetch all events and produce a deduplicated catalog showing unique event
names with aggregated counts.

Examples:
  sgx events catalog
  sgx events catalog --format table`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{"limit": {"100"}}
		data, _, err := c.GetAll(ctx, "/console/v1/events", params)
		if err != nil {
			return err
		}

		catalog := buildCatalog(data)
		catalogJSON, _ := json.Marshal(catalog)
		return output.PrintSuccess(getFormat(), "events.catalog", nil, json.RawMessage(catalogJSON), nil)
	},
}

func buildCatalog(raw []json.RawMessage) []catalogEntry {
	type event struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}

	byName := make(map[string]*catalogEntry)
	for _, item := range raw {
		var e event
		if json.Unmarshal(item, &e) != nil || e.Name == "" {
			continue
		}
		if existing, ok := byName[e.Name]; ok {
			existing.TotalCount += e.Count
			existing.Entries++
		} else {
			byName[e.Name] = &catalogEntry{
				Name:       e.Name,
				TotalCount: e.Count,
				Entries:    1,
			}
		}
	}

	result := make([]catalogEntry, 0, len(byName))
	for _, entry := range byName {
		result = append(result, *entry)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalCount > result[j].TotalCount
	})
	return result
}

func filterEventsByTime(data json.RawMessage, since, until string) (json.RawMessage, error) {
	if since == "" && until == "" {
		return data, nil
	}

	var events []map[string]any
	if err := json.Unmarshal(data, &events); err != nil {
		return data, nil
	}

	var sinceTime, untilTime time.Time
	if since != "" {
		t, err := time.Parse("2006-01-02", since)
		if err != nil {
			return nil, fmt.Errorf("invalid --since: %w", err)
		}
		sinceTime = t
	}
	if until != "" {
		t, err := time.Parse("2006-01-02", until)
		if err != nil {
			return nil, fmt.Errorf("invalid --until: %w", err)
		}
		untilTime = t.Add(24 * time.Hour)
	}

	filtered := make([]map[string]any, 0)
	for _, event := range events {
		dateStr, _ := event["date"].(string)
		if dateStr == "" {
			filtered = append(filtered, event)
			continue
		}
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			filtered = append(filtered, event)
			continue
		}
		if !sinceTime.IsZero() && t.Before(sinceTime) {
			continue
		}
		if !untilTime.IsZero() && t.After(untilTime) {
			continue
		}
		filtered = append(filtered, event)
	}

	result, _ := json.Marshal(filtered)
	return result, nil
}

func init() {
	eventsListCmd.Flags().StringVar(&eventsSince, "since", "", "Only events after date (YYYY-MM-DD)")
	eventsListCmd.Flags().StringVar(&eventsUntil, "until", "", "Only events before date (YYYY-MM-DD)")

	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsGetCmd)
	eventsCmd.AddCommand(eventsMetricsCmd)
	eventsCmd.AddCommand(eventsCatalogCmd)
	rootCmd.AddCommand(eventsCmd)
}
