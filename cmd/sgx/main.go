package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/nicolasacchi/sgx/internal/client"
	"github.com/nicolasacchi/sgx/internal/commands"
	"github.com/nicolasacchi/sgx/internal/output"
)

var version = "dev"

func main() {
	commands.SetVersion(version)
	if err := commands.Execute(); err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) {
			output.PrintError(output.FormatJSON, "", apiErr.Error(), apiErr.StatusCode)
			fmt.Fprintln(os.Stderr, "Error:", apiErr.Error())
			if apiErr.Hint != "" {
				fmt.Fprintln(os.Stderr, "Hint:", apiErr.Hint)
			}
			os.Exit(apiErr.ExitCode())
		}
		output.PrintError(output.FormatJSON, "", err.Error(), 0)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
