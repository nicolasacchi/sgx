# CLAUDE.md — sgx

Go CLI for Statsig. Single binary, readonly, JSON output, API key auth.

**API**: [Statsig Console API](https://docs.statsig.com/console-api/introduction) v20240601. Base URL: `https://statsigapi.net`.

## Authentication

Resolution order (first non-empty wins):

1. `--api-key` flag
2. `STATSIG_API_KEY` env var
3. `STATSIG_CONSOLE_KEY` env var
4. `~/.config/sgx/config.json` — project from `--project` flag, then `default_project`, then legacy `api_key`

Requires a **Console API Key** from [console.statsig.com/api_keys](https://console.statsig.com/api_keys).

### Multi-project config

```json
{
  "default_project": "production",
  "projects": {
    "production": { "api_key": "console-abc123" },
    "staging": { "api_key": "console-xyz789", "format": "table" }
  }
}
```

Each project can override `base_url` and `format`. Legacy flat config (`{"api_key": "..."}`) still works.

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` | — | Statsig Console API key (overrides env var) |
| `--project` | — | Use named project from config file |
| `--format` | `json` | Output format: `json`, `table`, `compact` |
| `--base-url` | `https://statsigapi.net` | API base URL |
| `--verbose` | false | Print request/response details to stderr |
| `--no-paginate` | false | Return first page only |
| `--page` | 0 | Specific page (disables auto-pagination) |
| `--limit` | 100 | Results per page (max 100) |

## Commands

### experiments

```bash
sgx experiments list                                          # List all experiments
sgx experiments list --status active                          # Filter by status
sgx experiments list --tags tag1,tag2 --creator john          # Filter by tags/creator
sgx experiments list --owner elian                            # Client-side owner filter
sgx experiments list --since 2025-02-20                       # Modified after date
sgx experiments get my_experiment                             # Get experiment details
sgx experiments context my_experiment                         # Get experiment context
sgx experiments pulse my_experiment                           # Auto-resolves control/test groups
sgx experiments pulse my_experiment --no-cuped --confidence 90
sgx experiments pulse my_experiment --control id1 --test id2  # Explicit group IDs
sgx experiments inspect my_experiment                         # Full parallel inspection (6 endpoints)
sgx experiments exposures my_experiment                       # Cumulative exposure counts
sgx experiments versions my_experiment                        # Version history
sgx experiments overrides my_experiment                       # Override rules
```

**experiments list flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--status` | — | `setup`, `active`, `decision_made`, `abandoned` |
| `--tags` | — | Comma-separated tag IDs |
| `--team` | — | Team ID |
| `--stale` | false | Only stale experiments |
| `--type` | — | Experiment type |
| `--created-after` | — | YYYY-MM-DD |
| `--created-before` | — | YYYY-MM-DD |
| `--creator` | — | Creator name (API-side) |
| `--owner` | — | Owner/creator substring match (client-side) |
| `--since` | — | Modified after date YYYY-MM-DD (client-side) |

**experiments pulse flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--cuped` | true | Apply CUPED variance reduction |
| `--no-cuped` | false | Disable CUPED |
| `--confidence` | 95 | Confidence interval 0-100 |
| `--date` | — | Specific date (YYYY-MM-DD) |
| `--bonferroni-variant` | false | Bonferroni correction per variant |
| `--bonferroni-metric` | false | Bonferroni correction per metric |
| `--bonferroni-weight` | 0 | Alpha allocated to primary metrics |
| `--bh-metric` | false | Benjamini-Hochberg per metric |
| `--bh-variant` | false | Benjamini-Hochberg per variant |
| `--control` | — | Control group ID |
| `--test` | — | Test group ID |

**experiments inspect**: Parallel fetch of experiment config, pulse (with auto-group-resolution), cumulative exposures, context, overrides, and versions. Returns merged JSON object with all 6 sections. Non-fatal failures for secondary data (pulse, exposures, etc.).

### gates

```bash
sgx gates list                                               # List all gates
sgx gates list --type STALE --include-archived               # Filter gates
sgx gates get my_gate                                        # Get gate details
sgx gates rules my_gate                                      # Get gate rules
sgx gates checks my_gate                                     # Get check counts
sgx gates pulse my_gate rule_123                             # Pulse for a specific rule
sgx gates pulse my_gate rule_123 --no-cuped --confidence 90
sgx gates refs my_gate                                       # All references (experiments, gates, configs)
```

**gates list flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--type` | — | `STALE`, `PERMANENT`, etc. |
| `--type-reason` | — | Type reason filter |
| `--tags` | — | Comma-separated tags |
| `--id-type` | — | ID type filter |
| `--include-archived` | false | Include archived gates |

**gates pulse flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--no-cuped` | false | Disable CUPED |
| `--confidence` | 95 | Confidence interval 0-100 |

### metrics

```bash
sgx metrics list                                             # List all metrics
sgx metrics list --show-hidden --tags core                   # Include hidden metrics
sgx metrics get add_to_cart::event_count                     # Get metric definition
sgx metrics value add_to_cart::event_count --date 2025-02-20 # Single metric value
sgx metrics values --date 2025-02-20 --name conversion_rate  # All metric values
sgx metrics experiments add_to_cart::event_count             # Experiments using metric
sgx metrics sql add_to_cart::event_count                     # SQL definition
sgx metrics sources                                          # List metric sources
```

Metric IDs use format `<name>::<type>` (e.g. `add_to_cart::event_count`).

**metrics list flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--tags` | — | Comma-separated tags |
| `--show-hidden` | false | Include hidden metrics |

**metrics value / values flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--date` | — | Specific date (YYYY-MM-DD) |
| `--name` | — | Filter by metric name (values only) |
| `--type` | — | Filter by metric type (values only) |

### events

```bash
sgx events list                                              # List logged events (default limit 50)
sgx events list --limit 20                                   # Custom limit
sgx events list --since 2025-02-20 --until 2025-02-25       # Time-filtered (client-side)
sgx events get purchase                                      # Get event details
sgx events metrics purchase                                  # Metrics derived from event
sgx events catalog                                           # Deduplicated event names with counts
sgx events catalog --format table                            # Tabular catalog
```

### holdouts

```bash
sgx holdouts list                                            # List all holdouts
sgx holdouts get my_holdout                                  # Get holdout details
sgx holdouts pulse my_holdout                                # Holdout pulse results
sgx holdouts pulse my_holdout --no-cuped --confidence 90
```

### layers

```bash
sgx layers list                                              # List all layers
sgx layers get my_layer                                      # Get layer details
sgx layers experiments my_layer                              # Experiments in layer
```

### exposures

```bash
sgx exposures                                                # All exposure counts
sgx exposures --experiments exp1,exp2                        # Filter by experiments
sgx exposures --gates gate1,gate2                            # Filter by gates
sgx exposures --configs config1,config2                      # Filter by dynamic configs
```

### reports

```bash
sgx reports --type pulse_daily --date 2025-02-20             # Download report URL
sgx reports --type first_exposures --date 2025-02-20
sgx reports --type topline_impact_daily --date 2025-02-20
```

Both `--type` and `--date` are required.

### audit

```bash
sgx audit                                                    # Recent audit logs
sgx audit --action experiment_start --start-date 2025-02-01
sgx audit --start-date 2025-02-01 --end-date 2025-02-25 --order asc
sgx audit --summary --start-date 2025-02-01                  # Grouped by day/user
```

| Flag | Default | Description |
|------|---------|-------------|
| `--action` | — | Action type (e.g. `experiment_start`, `gate_create`) |
| `--start-date` | — | YYYY-MM-DD |
| `--end-date` | — | YYYY-MM-DD |
| `--sort` | `date` | Sort key |
| `--order` | `desc` | `asc` or `desc` |
| `--summary` | false | Group entries by day and user |

### segments

```bash
sgx segments list                                            # List all segments
sgx segments get my_segment                                  # Get segment details
```

### overview

Aggregated project dashboard — fetches all resources in parallel.

```bash
sgx overview                                                 # Quick snapshot (pulse for first 10 experiments)
sgx overview --full                                          # Pulse for ALL active experiments
sgx overview --concurrency 3                                 # Limit parallel requests
sgx overview --experiments exp1,exp2                         # Pulse only for specific IDs
```

| Flag | Default | Description |
|------|---------|-------------|
| `--full` | false | Fetch pulse for all active experiments |
| `--concurrency` | 5 | Max parallel API requests |
| `--experiments` | — | Comma-separated experiment IDs |

Returns: `project_summary` (counts), `experiments` (with pulse data), `gates`, `stale_gates`, `holdouts`, `exposure_counts`, `alerts`.

### config

```bash
sgx config add production --api-key console-abc123           # Add/update a project
sgx config add staging --api-key console-xyz --format table  # With custom format
sgx config remove staging                                    # Remove a project
sgx config list                                              # List all projects (keys masked)
sgx config use production                                    # Set default project
sgx config current                                           # Show resolved active project
```

### version

```bash
sgx version                                                  # Version, Go version, OS, arch
```

## Output Format

All commands output JSON to stdout. Errors and diagnostics go to stderr.

**Success envelope:**

```json
{
  "ok": true,
  "command": "experiments.list",
  "args": {"status": "active"},
  "data": [...],
  "pagination": {"itemsPerPage": 100, "pageNumber": 1, "totalItems": 12, "nextPage": null}
}
```

**Error envelope:**

```json
{
  "ok": false,
  "command": "",
  "error": "404: Experiment not found.",
  "status_code": 404
}
```

**Table format** (available for: experiments.list, experiments.get, gates.list, metrics.list, events.list, events.catalog, holdouts.list, layers.list, segments.list, config.list, audit.summary, overview):

```bash
sgx experiments list --format table
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | API/network error |
| 2 | Auth error (401/403) |
| 3 | Invalid arguments (cobra) |
| 4 | Not found (404) |

## HTTP Client

| Setting | Value |
|---------|-------|
| Timeout | 30s per request |
| Retries | Max 3 on 429 or network error |
| Backoff | Exponential (1s, 2s, 4s) + 0-500ms jitter |
| Pagination | Auto-follows `nextPage`, merges data arrays, cap 20 pages |
| Headers | `STATSIG-API-KEY`, `STATSIG-API-VERSION: 20240601` |

## Build

```bash
make install                    # Install to $GOPATH/bin/sgx
make build                      # Build to ./bin/sgx
go install ./cmd/sgx            # Direct Go install
make test                       # Run tests
```

Requires Go 1.25+.

## Project Structure

```
cmd/sgx/main.go                          # Entry point, version injection, exit codes
internal/client/client.go                # HTTP client, retries, pagination
internal/client/client_test.go           # Client tests (httptest)
internal/config/config.go                # Auth resolution, multi-project config, CRUD
internal/config/config_test.go           # Config tests (project resolution, add/remove)
internal/commands/root.go                # Root command, global flags, getClient()
internal/commands/config.go              # config add, remove, list, use, current
internal/commands/segments.go            # segments list, get
internal/commands/experiments_inspect.go # experiments inspect (parallel 6-endpoint fetch)
internal/commands/experiments.go         # experiments list, get, context
internal/commands/experiments_pulse.go   # experiments pulse
internal/commands/experiments_extra.go   # experiments exposures, versions, overrides
internal/commands/gates.go               # gates list, get, rules, checks
internal/commands/gates_pulse.go         # gates pulse
internal/commands/gates_refs.go          # gates refs (merged references)
internal/commands/metrics.go             # metrics list, get, value, values, experiments, sql, sources
internal/commands/events.go              # events list, get, metrics
internal/commands/holdouts.go            # holdouts list, get, pulse
internal/commands/layers.go              # layers list, get, experiments
internal/commands/exposures.go           # exposure_count
internal/commands/reports.go             # report downloads
internal/commands/audit.go               # audit_logs
internal/commands/overview.go            # parallel aggregation (errgroup)
internal/commands/version.go             # version info
internal/commands/helpers.go             # mergeRawMessages helper
internal/output/envelope.go              # SuccessEnvelope, ErrorEnvelope
internal/output/output.go               # JSON/table/compact dispatcher
internal/output/table.go                 # go-pretty table renderer, column defs
```
