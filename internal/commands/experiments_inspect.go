package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var experimentsInspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Full experiment inspection (parallel fetch)",
	Long: `Fetch experiment details, pulse, exposures, context, overrides, and versions
in parallel and merge into a single response.

Examples:
  sgx experiments inspect my_experiment
  sgx experiments inspect my_experiment --format table`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentInspect,
}

type inspectData struct {
	mu        sync.Mutex
	config    json.RawMessage
	pulse     json.RawMessage
	exposures json.RawMessage
	context   json.RawMessage
	overrides json.RawMessage
	versions  json.RawMessage
}

func runExperimentInspect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	c, err := getClient(cmd)
	if err != nil {
		return err
	}

	expID := args[0]
	data := &inspectData{}
	g, gctx := errgroup.WithContext(ctx)

	// 1. Experiment config (fatal)
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/experiments/"+expID, nil)
		if err != nil {
			return fmt.Errorf("experiment get: %w", err)
		}
		data.config = resp.Data
		return nil
	})

	// 2. Pulse results (non-fatal — may need groups)
	g.Go(func() error {
		params := url.Values{"cuped": {"true"}, "confidence": {"95"}}
		resp, err := c.Get(gctx, "/console/v1/experiments/"+expID+"/pulse_results", params)
		if err != nil {
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "pulse: %v (will retry with auto-resolved groups)\n", err)
			}
			return nil
		}
		data.mu.Lock()
		data.pulse = resp.Data
		data.mu.Unlock()
		return nil
	})

	// 3. Cumulative exposures (non-fatal)
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/experiments/"+expID+"/cumulative_exposures", nil)
		if err != nil {
			return nil
		}
		data.mu.Lock()
		data.exposures = resp.Data
		data.mu.Unlock()
		return nil
	})

	// 4. Context (non-fatal)
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/experiments/"+expID+"/context", nil)
		if err != nil {
			return nil
		}
		data.mu.Lock()
		data.context = resp.Data
		data.mu.Unlock()
		return nil
	})

	// 5. Overrides (non-fatal)
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/experiments/"+expID+"/overrides", nil)
		if err != nil {
			return nil
		}
		data.mu.Lock()
		data.overrides = resp.Data
		data.mu.Unlock()
		return nil
	})

	// 6. Versions (non-fatal)
	g.Go(func() error {
		resp, err := c.Get(gctx, "/console/v1/experiments/"+expID+"/versions", nil)
		if err != nil {
			return nil
		}
		data.mu.Lock()
		data.versions = resp.Data
		data.mu.Unlock()
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	// Retry pulse with auto-resolved groups if initial fetch failed
	if data.pulse == nil {
		control, test, resolveErr := resolveGroups(ctx, c, expID)
		if resolveErr == nil {
			params := url.Values{"cuped": {"true"}, "confidence": {"95"}, "control": {control}, "test": {test}}
			resp, err := c.Get(ctx, "/console/v1/experiments/"+expID+"/pulse_results", params)
			if err == nil {
				data.pulse = resp.Data
			}
		}
	}

	result := map[string]json.RawMessage{
		"experiment": data.config,
		"pulse":      orNull(data.pulse),
		"exposures":  orNull(data.exposures),
		"context":    orNull(data.context),
		"overrides":  orNull(data.overrides),
		"versions":   orNull(data.versions),
	}
	resultJSON, _ := json.Marshal(result)

	return output.PrintSuccess(getFormat(), "experiments.inspect", map[string]any{"id": expID}, resultJSON, nil)
}

func orNull(data json.RawMessage) json.RawMessage {
	if data == nil {
		return json.RawMessage("null")
	}
	return data
}

func init() {
	experimentsCmd.AddCommand(experimentsInspectCmd)
}
