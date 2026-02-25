package commands

import (
	"encoding/json"
	"runtime"

	"github.com/nicolasacchi/sgx/internal/output"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		info := map[string]string{
			"version":    version,
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
		}

		data, _ := json.Marshal(info)
		return output.PrintSuccess(getFormat(), "version", nil, data, nil)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
