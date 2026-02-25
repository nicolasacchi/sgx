package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var gatesRefsCmd = &cobra.Command{
	Use:   "refs <id>",
	Short: "Get all references for a gate",
	Long: `Get experiment, gate, and dynamic config references for a gate.
Merges results from three endpoints into a single response.

Examples:
  sgx gates refs my_gate`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		gateID := args[0]
		basePath := "/console/v1/gates/" + gateID

		type refResult struct {
			data json.RawMessage
			err  error
		}

		expCh := make(chan refResult, 1)
		gateCh := make(chan refResult, 1)
		cfgCh := make(chan refResult, 1)

		go func() {
			resp, err := c.Get(ctx, basePath+"/experiment_references", nil)
			if err != nil {
				expCh <- refResult{err: err}
				return
			}
			expCh <- refResult{data: resp.Data}
		}()

		go func() {
			resp, err := c.Get(ctx, basePath+"/gate_references", nil)
			if err != nil {
				gateCh <- refResult{err: err}
				return
			}
			gateCh <- refResult{data: resp.Data}
		}()

		go func() {
			resp, err := c.Get(ctx, basePath+"/dynamic_config_references", nil)
			if err != nil {
				cfgCh <- refResult{err: err}
				return
			}
			cfgCh <- refResult{data: resp.Data}
		}()

		expRes := <-expCh
		gateRes := <-gateCh
		cfgRes := <-cfgCh

		// Collect errors
		var errs []error
		if expRes.err != nil {
			errs = append(errs, fmt.Errorf("experiment_references: %w", expRes.err))
		}
		if gateRes.err != nil {
			errs = append(errs, fmt.Errorf("gate_references: %w", gateRes.err))
		}
		if cfgRes.err != nil {
			errs = append(errs, fmt.Errorf("dynamic_config_references: %w", cfgRes.err))
		}
		if len(errs) == 3 {
			return errs[0]
		}

		merged := map[string]json.RawMessage{
			"experiments":     orEmpty(expRes.data),
			"gates":           orEmpty(gateRes.data),
			"dynamic_configs": orEmpty(cfgRes.data),
		}
		data, _ := json.Marshal(merged)

		return output.PrintSuccess(getFormat(), "gates.refs", map[string]any{"id": gateID}, data, nil)
	},
}

func orEmpty(data json.RawMessage) json.RawMessage {
	if data == nil {
		return json.RawMessage("[]")
	}
	return data
}

func init() {
	gatesCmd.AddCommand(gatesRefsCmd)
}
