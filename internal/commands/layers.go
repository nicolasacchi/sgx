package commands

import (
	"context"
	"fmt"
	"net/url"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var layersCmd = &cobra.Command{
	Use:   "layers",
	Short: "List and inspect layers",
	Long: `List and inspect layers and their experiments.

Examples:
  sgx layers list
  sgx layers get my_layer
  sgx layers experiments my_layer`,
}

var layersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all layers",
	Long: `List all layers.

Examples:
  sgx layers list`,
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

		resp, err := c.Get(ctx, "/console/v1/layers", params)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "layers.list", nil, resp.Data, resp.Pagination)
	},
}

var layersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get layer details",
	Long: `Get full details for a specific layer.

Examples:
  sgx layers get my_layer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/layers/"+args[0], nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "layers.get", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

var layersExperimentsCmd = &cobra.Command{
	Use:   "experiments <id>",
	Short: "Get experiments in this layer",
	Long: `Get all experiments that belong to a specific layer.

Examples:
  sgx layers experiments my_layer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		resp, err := c.Get(ctx, "/console/v1/layers/"+args[0]+"/experiments", nil)
		if err != nil {
			return err
		}
		return output.PrintSuccess(getFormat(), "layers.experiments", map[string]any{"id": args[0]}, resp.Data, nil)
	},
}

func init() {
	layersCmd.AddCommand(layersListCmd)
	layersCmd.AddCommand(layersGetCmd)
	layersCmd.AddCommand(layersExperimentsCmd)
	rootCmd.AddCommand(layersCmd)
}
