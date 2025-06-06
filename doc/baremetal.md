## `baremetal` Command

Manage dedicated **Bare Metal servers** – perform lifecycle actions, query hardware details, manage network interfaces, and trigger maintenance operations.

### Usage

```bash
ovhcloud baremetal [command] [flags]
```

The command inherits all **global flags** (`--debug`, `--json`, `--yaml`, `--format`, `--interactive`, `--filter`) and adds its own `-h/--help` flag at each level.

### Sub‑commands

| Sub‑command            | Purpose                                                                                  |
| ---------------------- | ---------------------------------------------------------------------------------------- |
| **boot**               | Manage boot options (e.g., network/drive boot, preferred kernel).                        |
| **edit**               | Update the bare‑metal’s metadata (name, description, ...).                               |
| **get**                | Display detailed information about a specific server (hardware, datacenter, OS, status). |
| **ipmi**               | Control and query IPMI for remote power/actions & KVM console.                           |
| **list**               | List all Bare‑Metal services in your account.                                            |
| **list-compatible-os** | List operating systems that can be installed on the server.                              |
| **list-interventions** | Show past & scheduled maintenance interventions.                                         |
| **list-ips**           | Display all IP blocks routed to the server.                                              |
| **list-secrets**       | Retrieve sensitive connection secrets (e.g., rescue passwords).                          |
| **list-tasks**         | List tasks & their progress for the server (installations, reboots, etc.).               |
| **reboot**             | Perform a hard reboot.                                                                   |
| **reboot-rescue**      | Reboot into OVHcloud rescue mode.                                                        |
| **reinstall**          | Reinstall the server with a chosen OS template.                                          |
| **vni**                | Manage **Virtual Network Interfaces** attached to vRack.                                 |

The commands `reboot-rescue` and `reinstall` can be used with a `--wait` flag that will make the CLI wait for the task to be completed and display the new authentication secrets afterwards.

### Examples

| Task                                        | Command                                                                          |
| ------------------------------------------- | -------------------------------------------------------------------------------- |
| List all servers                            | `ovhcloud baremetal list`                                                         |
| Get full details                            | `ovhcloud baremetal get <service_id>`                                             |
| Get full details (JSON)                     | `ovhcloud baremetal get <service_id> --json`                                      |
| Reboot a server                             | `ovhcloud baremetal reboot <service_id>`                                          |
| Reboot into rescue mode and notify by email | `ovhcloud baremetal reboot-rescue <service_id>`                                   |
| Check compatible OS templates               | `ovhcloud baremetal list-compatible-os <service_id>`                              |
| Open an IPMI KVM console                    | `ovhcloud baremetal ipmi get-access <service_id> --type serialOverLanURL --ttl 5` |                            |

### Tips

* Use `--json` or `--yaml` to parse output in automation scripts.
* Combine `--format` with a [gval] expression to extract single fields, e.g., server name or status.
* Combine `--filter` with a [gval] expression on commands that perform a listing to filter results.
* Append `--debug` to any command to inspect underlying API calls for troubleshooting.

[gval]: https://github.com/PaesslerAG/gval

> For full details on each sub‑command, run `ovhcloud baremetal <sub‑command> --help`.
