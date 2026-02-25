package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type Format string

const (
	FormatJSON    Format = "json"
	FormatTable   Format = "table"
	FormatCompact Format = "compact"
)

func PrintSuccess(format Format, command string, args map[string]any, data json.RawMessage, pagination any) error {
	env := SuccessEnvelope{
		OK:         true,
		Command:    command,
		Args:       args,
		Data:       data,
		Pagination: pagination,
	}

	switch format {
	case FormatTable:
		return printTable(command, data)
	case FormatCompact:
		return printCompactJSON(env)
	default:
		return printJSON(env)
	}
}

func PrintError(format Format, command string, errMsg string, statusCode int) {
	env := ErrorEnvelope{
		OK:         false,
		Command:    command,
		Error:      errMsg,
		StatusCode: statusCode,
	}
	enc := json.NewEncoder(os.Stdout)
	if format != FormatCompact {
		enc.SetIndent("", "  ")
	}
	enc.Encode(env)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

func printCompactJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}
