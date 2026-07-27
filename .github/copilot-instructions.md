# Copilot / AI agent instructions

Guidance for AI coding tools working in this repository (GitHub Copilot, Claude, …).
`CLAUDE.md` at the repo root is a symlink to this file, so every tool reads the same source.
Keep it accurate: if you change a convention or helper, update this file.

## What this is

`ovhcloud` — a Go CLI (cobra) wrapping the OVHcloud public APIs (v1 and v2). Each API
"universe" (cloud, domain, baremetal, vps, me, …) is exposed as a top-level command.
Module path: `github.com/ovh/ovhcloud-cli`. Entry point: `cmd/ovhcloud/main.go` → `cmd.Execute()`.

## Commands (always run before handing back)

```bash
make fmt          # gofmt — mandatory
make build        # CGO_ENABLED=0 build of ./cmd/ovhcloud → ./ovhcloud
go test ./...     # all tests must pass
make doc          # regenerate doc/ (see Docs below)
```

Refresh a **v1** OpenAPI schema: `make schemas UNIVERSE=<name>` (e.g. `cloud`, `domain`, `vps`).
There is **no** automated refresh for v2 schemas (see "API schemas" below).

## Architecture — the two-file pattern

Adding or changing a command almost always touches exactly two files:

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Command** | `internal/cmd/<universe>.go` | cobra definitions, flags, arg validation. **No business logic.** |
| **Service** | `internal/services/<universe>/*.go` | HTTP calls + response handling. **No printing.** |
| Display | `internal/display/` | ALL output/formatting (JSON/YAML/interactive/custom). |
| OpenAPI | `internal/openapi/` | reads embedded schemas to build request-body skeletons and filter editable fields. |
| Assets | `internal/assets/api-schemas/*.json` | embedded OpenAPI schemas (`//go:embed`). |

### Command registration

Each `internal/cmd/<universe>.go` has a `func init()` that builds its command tree and ends
with `rootCmd.AddCommand(<universe>Cmd)`. Cobra wires everything at startup — there is no
central registry to edit.

**Cloud is the exception**: `internal/cmd/cloud_project.go` `init()` builds `cloudCmd`, then calls
every `initCloud<Feature>Command(cloudCmd)`, and finally `rootCmd.AddCommand(cloudCmd)`. A new cloud
feature = a new `initCloud<Feature>Command` **that you must add to that block**. Storage sub-commands
are wired one level deeper in `internal/cmd/cloud_storage.go`.

Reference examples to copy from:
- v1 resource with nested sub-commands: `internal/services/cloud/cloud_storage_block.go` + `internal/cmd/cloud_storage_block.go`
- **v2** resource (list/get/create/edit/delete + action): `internal/services/cloud/cloud_managed_rancher.go` + `internal/cmd/cloud_managed_rancher.go`

## Core helpers — prefer these over hand-rolling

Service layer (`internal/services/common`):
- `common.ManageListRequestNoExpand(endpoint, columns, flags.GenericFilters)` — list.
- `common.ManageObjectRequest(endpoint, id, template)` — get (renders a `.tmpl`).
- `common.CreateResource(cmd, schemaPath, endpoint, example, spec, schemaBytes, requiredFields)` — create.
- `common.EditResource(cmd, schemaPath, endpoint, spec, schemaBytes)` — edit.
- `httpLib.Client.Get/Post/Delete(...)` — raw calls for simple actions only.
- `getConfiguredCloudProject()` (cloud package) — resolve the target cloud project.

Command layer (`internal/cmd/`):
- `withFilterFlag(cmd)` — adds `--filter` to a list command.
- `addParameterFileFlags(cmd, skipInit, schemaBytes, path, method, defaultExample, replaceFn)` — adds `--from-file` and `--init-file`.
- `addInteractiveEditorFlag(cmd)` — adds `--editor`.
- `markFlagsMutuallyExclusive(cmd, "from-file", "editor")`.

Output (`internal/display`):
- `display.OutputInfo(&flags.OutputFormatConfig, details, message, params...)` — success/info.
- `display.OutputError(&flags.OutputFormatConfig, message, params...)` — errors.
- **Never** use `fmt.Print*` to talk to the user. Always go through `display`.

## Conventions (enforced in review)

- **Name commands after the API endpoint.**
- **Output only via `internal/display`** — never `fmt.Println`.
- **`url.PathEscape` every identifier** injected into a URL path.
- **List column order: `id`, `name`, `region`, `type` first** when those fields exist.
  Column mapping syntax is `"jsonPath alias"`, e.g. `"currentState.name name"`.
- **Declare a cobra flag for every field** of a create/edit `Spec` struct (kebab-case flag names).
- **Create/edit UX rule** (CONTRIBUTING.md): if the request body has **> 5 parameters** or **more
  than one level of nesting**, the command MUST offer `--editor`, `--from-file`, and `--init-file`
  (via the helpers above).
- **`--editor`/`--from-file` must stay wired**: these flags only take effect through
  `common.CreateResource`/`EditResource` (they set globals read nowhere else). If a create/edit
  handler builds the body by hand and calls `httpLib.Client` directly, those flags become silent
  no-ops — route through `CreateResource`/`EditResource` instead, or drop the flags.
- **Async polling**: when waiting on a task/operation, a task/sub-resource reported in `ERROR` is
  **logged and waiting continues** (transient errors resolve backend-side); only a top-level
  resource/operation error status is fatal. See `internal/services/cloud/utils.go`.
- Keep changes small and single-purpose; avoid commands that fan out into many HTTP calls.

## API schemas — what the JSON files are for

`internal/assets/api-schemas/*.json` are embedded OpenAPI specs. They are **not** used for routing
(endpoints are hardcoded Go strings). They are read at runtime, via `internal/openapi`, only to:
1. **Build the request-body skeleton** for `--init-file` / `--editor` (`GetOperationRequestExamples`).
2. **Filter editable/unknown fields** before an edit is sent (`FilterEditableFields` → `pruneUnknownFields`).

So when the API contract changes, the corresponding `Spec` struct + cobra flags must be updated,
or users can't drive the new/changed fields.

**v1** (`cloud.json`, `me.json`, …): full spec minus `x-code-samples`, refreshed with
`make schemas UNIVERSE=<name>`.

**v2** (`cloud_v2.json`): a **hand-curated subset** — only the paths the CLI actually exposes plus
the schemas those paths reference (transitively) and OVH's standard scalar types. There is no `make`
target; it is maintained manually (only public — alpha/beta/stable — paths are curated in, never
`Internal use only` ones).

**Gotcha**: in the schema, v2 paths have **no `/v2` prefix** (`/publicCloud/project/{projectId}/rancher`),
but the Go HTTP calls and the `schemaPath` argument to `Create/EditResource` also omit `/v2` while the
actual `httpLib.Client` URL uses `/v2/publicCloud/...`. Match the surrounding code.

## Tests

Location: `internal/cmd/<universe>_test.go`. Pattern (see `cloud_storage_block_test.go`):
- `package cmd_test`; methods `func (ms *MockSuite) Test...(assert, require *td.T)`.
- Mock HTTP with `httpmock.RegisterResponder(method, "https://eu.api.ovh.com/v1|v2/...", httpmock.NewStringResponder(200, `<json>`))`.
- Drive the command with `cmd.Execute("cloud", "...", "--cloud-project", "fakeProjectID")`.
- Assert with go-testdeep: `require.CmpNoError(err)`, `assert.Cmp(out, td.Contains("..."))`.
- Cover the happy path, flag parsing, and at least one error path.

## Docs

`make doc` regenerates `doc/`. **Do not commit changes to `doc/ovhcloud.md`** unless they are
deliberate manual edits (the target checks it back out for that reason).

## Commit / PR

DCO required: sign commits with `Signed-off-by: Name <email>`. Keep PRs small and consistent with
existing patterns.

## Review checklist (apply when reviewing a PR — incl. GitHub Copilot code review)

Rank findings most-severe first; separate blockers (build/test failure, wrong behavior, missing
wiring) from nits (naming, ordering, style).

1. **Output layering** — user output via `internal/display` (`OutputInfo`/`OutputError`), never `fmt.Print*`.
2. **Command/service separation** — no HTTP in `internal/cmd`; no printing in `internal/services`.
3. **List column order** — `id, name, region, type` first when present; mapping syntax `"jsonPath alias"`.
4. **Create/edit UX flags** — body with > 5 params or nesting must offer `--editor`/`--from-file`/`--init-file`.
5. **`--editor`/`--from-file` wiring** — create/edit that bypasses `CreateResource`/`EditResource` makes these flags silent no-ops → flag it.
6. **URL safety** — `url.PathEscape` on every identifier in a URL path.
7. **Async error handling** — a task/sub-resource in `ERROR` while polling is logged and waiting continues; only a top-level resource/operation error is fatal.
8. **Command wiring** — new commands actually registered (`rootCmd.AddCommand`, or the `initCloud...Command` block in `cloud_project.go`).
9. **Flag coverage** — every `Spec` field has a kebab-case flag.
10. **Naming** — commands named after the API endpoint.
11. **Tests** — new/changed commands have `cmd_test.go` coverage (happy + error path); mocks target the correct `/v1` or `/v2` URL.
12. **v2 `/v2` gotcha** — `schemaPath` and Go paths omit `/v2` while the HTTP URL uses it; the embedded v2 schema is a curated public subset (don't suggest dumping the full spec).
13. **Docs** — if commands changed, `make doc` was run; `doc/ovhcloud.md` not committed unless a deliberate manual change.
14. **Build/tests green** — `make build` and `go test ./...` must pass (blocker if not).
