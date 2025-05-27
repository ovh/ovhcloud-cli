# OVHcloud CLI (`ovh-cli`) Documentation

---

## Overview

`ovh-cli` is a single, unified command‑line interface for managing the full range of OVHcloud products and account resources directly from your terminal. Whether you need to automate provisioning, perform quick look‑ups, or integrate OVHcloud operations into CI/CD pipelines, `ovh-cli` offers fine‑grained commands and consistent output formats (table, JSON, YAML, or custom gval expressions).

---

## Quick Start

```bash
# Display the top‑level help
ovh-cli --help

# Log in and create API credentials (interactive)
ovh-cli login

# List your VPS instances as JSON
ohv-cli vps list --json
```

### Generate Shell Completion

```bash
# Bash
eval "$(ovh-cli completion bash)"
# Zsh
eval "$(ovh-cli completion zsh)"
# Fish
ovh-cli completion fish | source
# PowerShell
ovh-cli completion powershell | Out-String | Invoke-Expression
```

Add the appropriate line to your shell’s startup file (`~/.bashrc`, `~/.zshrc`, etc.) to enable persistent autocompletion.

---

## Global Usage

```text
ovh-cli [command] [flags]
```

### Global Flags

| Flag              | Description                                          |
| ----------------- | ---------------------------------------------------- |
| `--debug`         | Activate debug mode (logs all HTTP‑request details). |
| `--format <expr>` | Format output with a [gval] expression.              |
| `--filter <expr>` | Filter lists output with a [gval] expression.        |
| `-h`, `--help`    | Display help for `ovh-cli` or a specific command.    |
| `--interactive`   | Produce interactive (prompt‑based) output.           |
| `--json`          | Output data in JSON format.                          |
| `--yaml`          | Output data in YAML format.                          |

[gval]: https://github.com/PaesslerAG/gval

#### Filtering examples

- Strict string equality: `--filter 'name=="something"'`
- String regexp comparison: `--filter 'name=~"something"'`
- Number comparison: `--filter 'bootId > 1'`

#### Formatting example

- Extract only one field: `--format 'ip'`
- Extract an object: `--format '{name: ip}'`

---

## Command Reference

Below is the full list of primary sub‑commands available at the time of writing. Each can be explored in depth with `ovh-cli <command> --help`.

| Command                              | Purpose                                                        |
| ------------------------------------ | -------------------------------------------------------------- |
| **alldom**                           | Retrieve information and manage your **AllDom** services.      |
| [**baremetal**](baremetal.md)        | Retrieve information and manage your **Bare Metal** servers.   |
| **cdn-dedicated**                    | Manage your dedicated **CDN** services.                        |
| **cloud**                            | Manage Public Cloud projects and services.                     |
| **completion**                       | Generate autocompletion scripts.                               |
| **config**                           | View and edit local CLI configuration (endpoints, auth, etc.). |
| **dedicated-ceph**                   | Manage **Dedicated Ceph** clusters.                            |
| **dedicated-cloud**                  | Manage **Dedicated Cloud** services.                           |
| **dedicated-cluster**                | Manage **Dedicated Cluster** resources.                        |
| **dedicated-nasha**                  | Manage **Nas‑HA** (High‑Availability NAS) services.            |
| **domain-name**                      | Manage domain names.                                           |
| **domain-zone**                      | Manage DNS zones.                                              |
| **email-domain**                     | Manage **Email Domain** services.                              |
| **email-mxplan**                     | Manage **Email MX Plan** services.                             |
| **email-pro**                        | Manage **Email Pro** services.                                 |
| **help**                             | Show help for any command.                                     |
| **hosting-private-database**         | Manage **Private SQL Database** hosting.                       |
| **iam**                              | Manage IAM resources, permissions, and policies.               |
| **ip**                               | Manage IP services (fail‑over, blocks, etc.).                  |
| **iploadbalancing**                  | Manage IP Load Balancing services.                             |
| **ldp**                              | Manage **Link Data Platform** services.                        |
| **location**                         | Look up datacenter and region locations.                       |
| **login**                            | Interactive login and API‑credential creation.                 |
| **nutanix**                          | Manage **Nutanix** environments.                               |
| **okms**                             | Manage **OVHcloud Key Management Service**.                    |
| **overthebox**                       | Manage **Over‑The‑Box** services.                              |
| **ovhcloudconnect**                  | Manage **OVHcloud Connect** links.                             |
| **pack-xdsl**                        | Manage **Pack xDSL**.                                          |
| **sms**                              | Manage **SMS** messaging services.                             |
| **ssl**                              | Manage SSL certificates across products.                       |
| **ssl-gateway**                      | Manage **SSL Gateway** services.                               |
| **storage-netapp**                   | Manage **NetApp** storage volumes.                             |
| **support-tickets**                  | List and manage support tickets.                               |
| **telephony**                        | Manage **Telephony** lines and services.                       |
| **veeamcloudconnect**                | Manage **Veeam Cloud Connect** backups.                        |
| **veeamenterprise**                  | Manage **Veeam Enterprise** services.                          |
| **vmwareclouddirector-backup**       | Manage **VMware Cloud Director** backups.                      |
| **vmwareclouddirector-organization** | Manage **VMware Cloud Director** organizations.                |
| **vps**                              | Manage **VPS** instances.                                      |
| **vrack**                            | Manage **vRack** networking.                                   |
| **vrackservices**                    | Manage **vRack Services**.                                     |
| **webhosting**                       | Manage **Web Hosting** plans.                                  |
| **xdsl**                             | Manage standalone **xDSL** lines.                              |

> **Tip**  Use `--json`, `--yaml`, or `--format` with a gval expression to integrate `ovh-cli` into scripts and automation pipelines.

---

## Examples

| Task                                  | Command                                        |
| ------------------------------------- | ---------------------------------------------- |
| Log in and save credentials           | `ovh-cli login`                                |
| List VPS instances (tabular)          | `ovh-cli vps list`                             |
| Fetch details of a single VPS in JSON | `ovh-cli vps get <service_id> --json`          |
| Reinstall a baremetal interactively   | `ovh-cli baremetal reinstall <id> --editor`    |

---

## Troubleshooting

* **Verbose output** — Use `--debug` to inspect raw API calls and responses.
* **Authentication issues** — Run `ovh-cli login` again to regenerate valid API keys.
* **Rate limits** — OVHcloud APIs impose rate limits; plan retries or exponential backoff in scripts.

---

## Further Reading

* OVHcloud API reference: [https://eu.api.ovh.com/console](https://eu.api.ovh.com/console)
* OVHcloud community guides and tutorials.

---

*Documentation generated from built‑in `ovh-cli --help` output. Feel free to edit and extend as new commands or flags are released.*
