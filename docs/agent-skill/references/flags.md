# Reference: global flags

These flags work on (almost) every command:

| Flag | Purpose |
|------|---------|
| `-o, --output json\|yaml\|interactive` | Machine-readable output. Use `json` for scripting. |
| `--output '<expr>'` | Extract/transform fields with a gval expression, e.g. `--output 'id'`, `--output '[id,"name"]'`, `--output 'nested.field'`. |
| `--filter '<expr>'` | Filter `list` results, e.g. `--filter 'status=="running"'`, `--filter 'name=~"^web"'`. |
| `--wait` | For async `create`/actions: block until the resource is ready. |
| `-d, --debug` | Log the full HTTP request/response (best for troubleshooting). |
| `--profile <name>` | Use a specific profile from the configuration (multi-account). |
| `-e, --ignore-errors` | Don't fail the whole command on a non-fatal per-item error. |

Notes:
- `--output '<expr>'` currently prints strings JSON-quoted; strip with `jq -r`
  or `tr -d '"'` if you need the bare value in a shell variable.
- `--filter`/`--output` use the gval syntax
  (https://github.com/PaesslerAG/gval).
