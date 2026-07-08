# CLAUDE.md

Guidance for AI agents working in this repository. Keep it accurate: if you change a
convention or helper, update this file.

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
every `initCloud<Feature>Command(cloudCmd)` (see the block around line 104), and finally
`rootCmd.AddCommand(cloudCmd)`. A new cloud feature = a new `initCloud<Feature>Command` **that you
must add to that block**. Storage sub-commands are wired one level deeper in `internal/cmd/cloud_storage.go`.

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

**v2** (`cloud_v2.json`): a **hand-curated subset** — only the paths the CLI actually exposes (14 of
~327) plus the schemas those paths reference (transitively) and OVH's standard scalar types. There is
no `make` target; it is maintained manually.

**Gotcha**: in the schema, v2 paths have **no `/v2` prefix** (`/publicCloud/project/{projectId}/rancher`),
but the Go HTTP calls and the `schemaPath` argument to `Create/EditResource` also omit `/v2` while the
actual `httpLib.Client` URL uses `/v2/publicCloud/...`. Match the surrounding code.

## Tests

Location: `internal/cmd/<universe>_test.go`. Pattern (see `cloud_storage_block_test.go`):
- `package cmd_test`; methods `func (ms *MockSuite) Test...(assert, require *td.T)`.
- Mock HTTP with `httpmock.RegisterResponder(method, "https://eu.api.ovh.com/v1|v2/...", httpmock.NewStringResponder(200, ` + "`json`" + `))`.
- Drive the command with `cmd.Execute("cloud", "...", "--cloud-project", "fakeProjectID")`.
- Assert with go-testdeep: `require.CmpNoError(err)`, `assert.Cmp(out, td.Contains("..."))`.
- Cover the happy path, flag parsing, and at least one error path.

## Docs

`make doc` regenerates `doc/`. **Do not commit changes to `doc/ovhcloud.md`** unless they are
deliberate manual edits (the target checks it back out for that reason).

## Commit / PR

DCO required: sign commits with `Signed-off-by: Name <email>`. Keep PRs small and consistent with
existing patterns.
