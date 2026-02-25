package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var (
	auditAction    string
	auditStartDate string
	auditEndDate   string
	auditSort      string
	auditOrder     string
	auditSummary   bool
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View audit logs",
	Long: `View audit trail of configuration changes.

Examples:
  sgx audit
  sgx audit --action experiment_start --start-date 2025-02-01
  sgx audit --start-date 2025-02-01 --end-date 2025-02-25 --order asc
  sgx audit --summary --start-date 2025-02-01`,
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

		var finalData json.RawMessage

		if noPaginate || pageFlag > 0 {
			resp, err := c.Get(ctx, "/console/v1/audit_logs", params)
			if err != nil {
				return err
			}
			if !auditSummary {
				return output.PrintSuccess(getFormat(), "audit", cmdArgs, resp.Data, resp.Pagination)
			}
			finalData = resp.Data
		} else {
			data, pagination, err := c.GetAll(ctx, "/console/v1/audit_logs", params)
			if err != nil {
				return err
			}
			if !auditSummary {
				return output.PrintSuccess(getFormat(), "audit", cmdArgs, mergeRawMessages(data), pagination)
			}
			finalData = mergeRawMessages(data)
		}

		summaryData := buildAuditSummary(finalData)
		summaryJSON, _ := json.Marshal(summaryData)
		return output.PrintSuccess(getFormat(), "audit.summary", cmdArgs, json.RawMessage(summaryJSON), nil)
	},
}

type auditDaySummary struct {
	Date  string             `json:"date"`
	Total int                `json:"total"`
	Users []auditUserSummary `json:"users"`
}

type auditUserSummary struct {
	User    string         `json:"user"`
	Actions map[string]int `json:"actions"`
	Count   int            `json:"count"`
}

func buildAuditSummary(data json.RawMessage) []auditDaySummary {
	var entries []map[string]any
	if json.Unmarshal(data, &entries) != nil {
		return nil
	}

	// Group by date -> user -> action
	type userActions struct {
		actions map[string]int
	}
	grouped := make(map[string]map[string]*userActions)

	for _, entry := range entries {
		dateStr, _ := entry["date"].(string)
		user, _ := entry["userName"].(string)
		action, _ := entry["actionType"].(string)
		if dateStr == "" {
			continue
		}
		if len(dateStr) > 10 {
			dateStr = dateStr[:10]
		}
		if user == "" {
			user = "(unknown)"
		}

		if grouped[dateStr] == nil {
			grouped[dateStr] = make(map[string]*userActions)
		}
		if grouped[dateStr][user] == nil {
			grouped[dateStr][user] = &userActions{actions: make(map[string]int)}
		}
		grouped[dateStr][user].actions[action]++
	}

	var result []auditDaySummary
	for date, users := range grouped {
		day := auditDaySummary{Date: date}
		for user, ua := range users {
			us := auditUserSummary{User: user, Actions: ua.actions}
			for _, count := range ua.actions {
				us.Count += count
			}
			day.Users = append(day.Users, us)
			day.Total += us.Count
		}
		sort.Slice(day.Users, func(i, j int) bool {
			return day.Users[i].Count > day.Users[j].Count
		})
		result = append(result, day)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date > result[j].Date
	})
	return result
}

func init() {
	auditCmd.Flags().StringVar(&auditAction, "action", "", "Filter by action type (e.g. experiment_start, gate_create)")
	auditCmd.Flags().StringVar(&auditStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	auditCmd.Flags().StringVar(&auditEndDate, "end-date", "", "End date (YYYY-MM-DD)")
	auditCmd.Flags().StringVar(&auditSort, "sort", "", "Sort key (default: date)")
	auditCmd.Flags().StringVar(&auditOrder, "order", "", "Sort order: asc or desc (default: desc)")
	auditCmd.Flags().BoolVar(&auditSummary, "summary", false, "Group entries by day and user")

	rootCmd.AddCommand(auditCmd)
}
