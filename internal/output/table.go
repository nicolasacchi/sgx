package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

type FormatFunc func(any) string

type ColumnDef struct {
	Header string
	Key    string
	Format FormatFunc
}

func countArray(v any) string {
	if arr, ok := v.([]any); ok {
		return fmt.Sprintf("%d", len(arr))
	}
	return "0"
}

func joinStrings(v any) string {
	if arr, ok := v.([]any); ok {
		parts := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

func formatGroups(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		g, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := g["name"].(string)
		size, _ := g["size"].(float64)
		parts = append(parts, fmt.Sprintf("%s(%g%%)", name, size))
	}
	return strings.Join(parts, ", ")
}

var commandColumns = map[string][]ColumnDef{
	"experiments.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "STATUS", Key: "status"},
		{Header: "TYPE", Key: "experimentType"},
		{Header: "ALLOC%", Key: "allocation"},
		{Header: "GROUPS", Key: "groups", Format: formatGroups},
		{Header: "TAGS", Key: "tags", Format: joinStrings},
	},
	"experiments.get": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "STATUS", Key: "status"},
		{Header: "HYPOTHESIS", Key: "hypothesis"},
		{Header: "ALLOC%", Key: "allocation"},
		{Header: "GROUPS", Key: "groups", Format: formatGroups},
	},
	"gates.list": {
		{Header: "ID", Key: "id"},
		{Header: "TYPE", Key: "type"},
		{Header: "ENABLED", Key: "isEnabled"},
		{Header: "RULES", Key: "rules", Format: countArray},
		{Header: "TAGS", Key: "tags", Format: joinStrings},
	},
	"metrics.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "TYPE", Key: "type"},
		{Header: "TAGS", Key: "tags", Format: joinStrings},
	},
	"events.list": {
		{Header: "NAME", Key: "name"},
		{Header: "COUNT", Key: "count"},
	},
	"holdouts.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
		{Header: "STATUS", Key: "status"},
	},
	"layers.list": {
		{Header: "ID", Key: "id"},
		{Header: "NAME", Key: "name"},
	},
}

func printTable(command string, data json.RawMessage) error {
	columns, ok := commandColumns[command]
	if !ok {
		// Fallback: just print as JSON
		return printJSON(SuccessEnvelope{OK: true, Command: command, Data: data})
	}

	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		// Try as single object
		var single map[string]any
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return printJSON(SuccessEnvelope{OK: true, Command: command, Data: data})
		}
		rows = []map[string]any{single}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if term.IsTerminal(int(os.Stdout.Fd())) {
		t.SetStyle(table.StyleLight)
	} else {
		t.SetStyle(table.StyleDefault)
	}

	header := make(table.Row, len(columns))
	for i, col := range columns {
		header[i] = col.Header
	}
	t.AppendHeader(header)

	for _, row := range rows {
		r := make(table.Row, len(columns))
		for i, col := range columns {
			if col.Format != nil {
				r[i] = col.Format(row[col.Key])
			} else {
				r[i] = formatValue(row[col.Key])
			}
		}
		t.AppendRow(r)
	}

	t.Render()
	return nil
}

func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
