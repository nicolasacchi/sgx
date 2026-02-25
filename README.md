# sgx

Readonly CLI for Statsig. Single binary, JSON output, API key auth.

Uses the [Statsig Console API](https://docs.statsig.com/console-api/introduction) to inspect experiments, gates, metrics, events, holdouts, layers, and more.

## Install

### From source

```bash
go install github.com/nicolasacchi/sgx/cmd/sgx@latest
```

### From release

Download the binary for your platform from [Releases](https://github.com/nicolasacchi/sgx/releases).

```bash
curl -L https://github.com/nicolasacchi/sgx/releases/latest/download/sgx_Linux_x86_64.tar.gz | tar xz
mv sgx ~/.local/bin/
```

### From source (local)

```bash
git clone https://github.com/nicolasacchi/sgx.git
cd sgx
make install
```

## Authentication

Set your Statsig Console API key via one of:

1. `--api-key` flag
2. `STATSIG_API_KEY` environment variable
3. `STATSIG_CONSOLE_KEY` environment variable
4. `~/.config/sgx/config.json` (named projects or legacy flat config)

Generate a Console API key at [console.statsig.com/api_keys](https://console.statsig.com/api_keys).

### Multi-project configuration

Store multiple API keys as named projects:

```bash
sgx config add production --api-key console-abc123
sgx config add staging --api-key console-xyz789
sgx config list
sgx config use production        # Set default project
sgx config current               # Show active project
```

Use `--project` to switch per-command:

```bash
sgx experiments list --project staging
```

Config file (`~/.config/sgx/config.json`):

```json
{
  "default_project": "production",
  "projects": {
    "production": { "api_key": "console-abc123" },
    "staging": { "api_key": "console-xyz789", "format": "table" }
  }
}
```

## Usage

### Experiments

```bash
sgx experiments list --status active
sgx experiments list --owner elian --since 2025-02-20
sgx experiments get my_experiment
sgx experiments pulse my_experiment                     # Auto-resolves groups
sgx experiments pulse my_experiment --confidence 95
sgx experiments inspect my_experiment                   # Full parallel inspection
sgx experiments exposures my_experiment
```

### Gates

```bash
sgx gates list
sgx gates get my_gate
sgx gates rules my_gate
sgx gates refs my_gate
```

### Metrics

```bash
sgx metrics list
sgx metrics get add_to_cart::event_count
sgx metrics value add_to_cart::event_count --date 2025-02-20
sgx metrics experiments add_to_cart::event_count
```

### Events

```bash
sgx events list
sgx events list --since 2025-02-20 --until 2025-02-25
sgx events catalog                                     # Deduplicated event names
sgx events catalog --format table
```

### Segments

```bash
sgx segments list
sgx segments get my_segment
```

### Overview (project snapshot)

```bash
sgx overview                    # Parallel fetch of all resources
sgx overview --full             # Include pulse for all active experiments
sgx overview --format table     # Human-readable dashboard
```

### Audit trail

```bash
sgx audit --start-date 2025-02-01 --end-date 2025-02-25
sgx audit --summary --start-date 2025-02-01             # Grouped by day/user
```

## Output

Three formats: `--format json` (default), `--format table`, `--format compact`.

### JSON (default)

```json
{
  "ok": true,
  "command": "experiments.list",
  "args": {"status": "active"},
  "data": [...],
  "pagination": {"itemsPerPage": 100, "pageNumber": 1, "totalItems": 12}
}
```

### Table

```
+----------------------------+-------------------------+--------+------+--------+-------------------------+------+
| ID                         | NAME                    | STATUS | TYPE | ALLOC% | GROUPS                  | TAGS |
+----------------------------+-------------------------+--------+------+--------+-------------------------+------+
| checkout_redesign          | Checkout Redesign       | active | BASE | 100    | Control(50%), Test(50%) |      |
+----------------------------+-------------------------+--------+------+--------+-------------------------+------+
```

### Compact (single-line JSON)

```bash
sgx experiments list --format compact | jq '.data | length'
```

### Errors

```json
{
  "ok": false,
  "command": "",
  "error": "404: Experiment not found.",
  "status_code": 404
}
```

Exit codes: 0 = success, 1 = API error, 2 = auth error, 4 = not found.

## License

MIT
