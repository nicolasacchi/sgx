package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/nicolasacchi/sgx/internal/client"
	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	overviewFull        bool
	overviewConcurrency int
	overviewExperiments string
)

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Aggregated project dashboard",
	Long: `Fetches experiments, gates, holdouts, exposure counts, and alerts in parallel
and produces a unified project snapshot. For active experiments, also fetches
pulse results (up to 10 by default, all with --full).

Examples:
  sgx overview
  sgx overview --full
  sgx overview --concurrency 3
  sgx overview --experiments exp1,exp2`,
	RunE: runOverview,
}

type overviewData struct {
	mu sync.Mutex

	activeExperiments json.RawMessage
	setupExperiments  json.RawMessage
	gates             json.RawMessage
	holdouts          json.RawMessage
	exposureCounts    json.RawMessage
	alerts            json.RawMessage
	pulseResults      map[string]json.RawMessage
}

func runOverview(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	c, err := getClient(cmd)
	if err != nil {
		return err
	}

	data := &overviewData{
		pulseResults: make(map[string]json.RawMessage),
	}

	g, gctx := errgroup.WithContext(ctx)

	if verboseFlag {
		fmt.Fprintln(os.Stderr, "fetching project overview...")
	}

	// Fetch active experiments
	g.Go(func() error {
		params := url.Values{"status": {"active"}, "limit": {"100"}}
		resp, err := c.Get(gctx, "/console/v1/experiments", params)
		if err != nil {
			return fmt.Errorf("active experiments: %w", err)
		}
		data.activeExperiments = resp.Data
		return nil
	})

	// Fetch setup experiments
	g.Go(func() error {
		params := url.Values{"status": {"setup"}, "limit": {"100"}}
		resp, err := c.Get(gctx, "/console/v1/experiments", params)
		if err != nil {
			return fmt.Errorf("setup experiments: %w", err)
		}
		data.setupExperiments = resp.Data
		return nil
	})

	// Fetch gates
	g.Go(func() error {
		params := url.Values{"limit": {"100"}}
		resp, err := c.Get(gctx, "/console/v1/gates", params)
		if err != nil {
			return fmt.Errorf("gates: %w", err)
		}
		data.gates = resp.Data
		return nil
	})

	// Fetch holdouts
	g.Go(func() error {
		params := url.Values{"limit": {"100"}}
		resp, err := c.Get(gctx, "/console/v1/holdouts", params)
		if err != nil {
			return fmt.Errorf("holdouts: %w", err)
		}
		data.holdouts = resp.Data
		return nil
	})

	// Fetch exposure counts
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/exposure_count", nil)
		if err != nil {
			return fmt.Errorf("exposure counts: %w", err)
		}
		data.exposureCounts = resp.Data
		return nil
	})

	// Fetch alerts
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/alerts", nil)
		if err != nil {
			// Alerts might not be available — don't fail
			data.alerts = json.RawMessage("[]")
			return nil
		}
		data.alerts = resp.Data
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	// Fetch pulse results for active experiments
	if err := fetchPulseResults(ctx, c, data); err != nil {
		if verboseFlag {
			fmt.Fprintf(os.Stderr, "warning: some pulse results failed: %v\n", err)
		}
	}

	// Assemble overview
	result := assembleOverview(data)
	resultJSON, _ := json.Marshal(result)

	return output.PrintSuccess(getFormat(), "overview", nil, resultJSON, nil)
}

func fetchPulseResults(ctx context.Context, c *client.Client, data *overviewData) error {
	var experiments []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data.activeExperiments, &experiments); err != nil {
		return nil // Not an array, skip
	}

	maxPulse := 10
	if overviewFull {
		maxPulse = len(experiments)
	}
	if maxPulse > len(experiments) {
		maxPulse = len(experiments)
	}

	// Semaphore for concurrency control
	sem := make(chan struct{}, overviewConcurrency)
	g, gctx := errgroup.WithContext(ctx)

	for i := 0; i < maxPulse; i++ {
		expID := experiments[i].ID
		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			params := url.Values{"cuped": {"true"}, "confidence": {"95"}}
			resp, err := c.Get(gctx, "/console/v1/experiments/"+expID+"/pulse_results", params)
			if err != nil {
				return nil // Non-fatal: skip failed pulse
			}

			data.mu.Lock()
			data.pulseResults[expID] = resp.Data
			data.mu.Unlock()
			return nil
		})
	}

	return g.Wait()
}

func assembleOverview(data *overviewData) map[string]any {
	var activeExps []map[string]any
	json.Unmarshal(data.activeExperiments, &activeExps)

	var setupExps []map[string]any
	json.Unmarshal(data.setupExperiments, &setupExps)

	var gates []map[string]any
	json.Unmarshal(data.gates, &gates)

	var holdouts []map[string]any
	json.Unmarshal(data.holdouts, &holdouts)

	// Count stale gates
	var staleGates []map[string]any
	for _, g := range gates {
		if gType, ok := g["type"].(string); ok && gType == "STALE" {
			staleGates = append(staleGates, g)
		}
	}

	var alerts []any
	json.Unmarshal(data.alerts, &alerts)

	// Enrich experiments with pulse results
	for i, exp := range activeExps {
		if id, ok := exp["id"].(string); ok {
			if pulse, found := data.pulseResults[id]; found {
				activeExps[i]["pulse"] = json.RawMessage(pulse)
			}
		}
	}

	summary := map[string]any{
		"active_experiments": len(activeExps),
		"setup_experiments":  len(setupExps),
		"total_gates":        len(gates),
		"stale_gates":        len(staleGates),
		"active_holdouts":    len(holdouts),
		"active_alerts":      len(alerts),
	}

	return map[string]any{
		"generated_at":    time.Now().UTC().Format(time.RFC3339),
		"project_summary": summary,
		"experiments":     activeExps,
		"gates":           gates,
		"stale_gates":     staleGates,
		"holdouts":        holdouts,
		"exposure_counts": json.RawMessage(data.exposureCounts),
		"alerts":          json.RawMessage(data.alerts),
	}
}

func init() {
	overviewCmd.Flags().BoolVar(&overviewFull, "full", false, "Fetch pulse for ALL active experiments (not just first 10)")
	overviewCmd.Flags().IntVar(&overviewConcurrency, "concurrency", 5, "Max parallel API requests for pulse fetching")
	overviewCmd.Flags().StringVar(&overviewExperiments, "experiments", "", "Only fetch pulse for these experiment IDs (comma-separated)")

	rootCmd.AddCommand(overviewCmd)
}
