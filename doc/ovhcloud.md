# OVHcloud CLI (`ovhcloud`) Documentation

---

## Overview

`ovhcloud` is a single, unified command‑line interface for managing the full range of OVHcloud products and account resources directly from your terminal. Whether you need to automate provisioning, perform quick look‑ups, or integrate OVHcloud operations into CI/CD pipelines, `ovhcloud` offers fine‑grained commands and consistent output formats (table, JSON, YAML, or custom gval expressions).

---

## Quick Start

```bash
# Display the top‑level help
ovhcloud --help

# Log in and create API credentials (interactive)
ovhcloud login

# List your VPS instances as JSON
ohvcloud vps list --json
```

Check out the [authentication page](authentication.md) for further information about the authentication means.

### Generate Shell Completion

```bash
# Bash
eval "$(./ovhcloud completion bash)"
# Zsh
eval "$(./ovhcloud completion zsh)"
# Fish
./ovhcloud completion fish | source
# PowerShell
./ovhcloud completion powershell | Out-String | Invoke-Expression
```

Add the appropriate line to your shell’s startup file (`~/.bashrc`, `~/.zshrc`, etc.) to enable persistent autocompletion.

---

## Global Usage

```text
ovhcloud [command] [flags]
```

### Global Flags

| Flag              | Description                                          |
| ----------------- | ---------------------------------------------------- |
| `--debug`         | Activate debug mode (logs all HTTP‑request details). |
| `--ignore-errors` | Ignore errors of API calls made when listing items.  |
| `--format <expr>` | Format output with a [gval] expression.              |
| `--filter <expr>` | Filter lists output with a [gval] expression.        |
| `-h`, `--help`    | Display help for `ovhcloud` or a specific command.   |
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

Below is the full list of primary sub‑commands available at the time of writing. Each can be explored in depth with `ovhcloud <command> --help`.

* [ovhcloud account](ovhcloud_account.md)	 - Manage your account
* [ovhcloud alldom](ovhcloud_alldom.md)	 - Retrieve information and manage your AllDom services
* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services
* [ovhcloud cdn-dedicated](ovhcloud_cdn-dedicated.md)	 - Retrieve information and manage your dedicated CDN services
* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud config](ovhcloud_config.md)	 - Manage your CLI configuration
* [ovhcloud dedicated-ceph](ovhcloud_dedicated-ceph.md)	 - Retrieve information and manage your Dedicated Ceph services
* [ovhcloud dedicated-cloud](ovhcloud_dedicated-cloud.md)	 - Retrieve information and manage your DedicatedCloud services
* [ovhcloud dedicated-cluster](ovhcloud_dedicated-cluster.md)	 - Retrieve information and manage your DedicatedCluster services
* [ovhcloud dedicated-nasha](ovhcloud_dedicated-nasha.md)	 - Retrieve information and manage your Dedicated NasHA services
* [ovhcloud domain-name](ovhcloud_domain-name.md)	 - Retrieve information and manage your domain names
* [ovhcloud domain-zone](ovhcloud_domain-zone.md)	 - Retrieve information and manage your domain zones
* [ovhcloud email-domain](ovhcloud_email-domain.md)	 - Retrieve information and manage your Email Domain services
* [ovhcloud email-mxplan](ovhcloud_email-mxplan.md)	 - Retrieve information and manage your Email MXPlan services
* [ovhcloud email-pro](ovhcloud_email-pro.md)	 - Retrieve information and manage your EmailPro services
* [ovhcloud hosting-private-database](ovhcloud_hosting-private-database.md)	 - Retrieve information and manage your HostingPrivateDatabase services
* [ovhcloud iam](ovhcloud_iam.md)	 - Manage IAM resources, permissions and policies
* [ovhcloud ip](ovhcloud_ip.md)	 - Retrieve information and manage your IP services
* [ovhcloud iploadbalancing](ovhcloud_iploadbalancing.md)	 - Retrieve information and manage your IP LoadBalancing services
* [ovhcloud ldp](ovhcloud_ldp.md)	 - Retrieve information and manage your LDP (Logs Data Platform) services
* [ovhcloud location](ovhcloud_location.md)	 - Retrieve information and manage your Location services
* [ovhcloud login](ovhcloud_login.md)	 - Login to your OVHcloud account to create API credentials
* [ovhcloud nutanix](ovhcloud_nutanix.md)	 - Retrieve information and manage your Nutanix services
* [ovhcloud okms](ovhcloud_okms.md)	 - Retrieve information and manage your OKMS (Key Management Services)
* [ovhcloud overthebox](ovhcloud_overthebox.md)	 - Retrieve information and manage your OverTheBox services
* [ovhcloud ovhcloudconnect](ovhcloud_ovhcloudconnect.md)	 - Retrieve information and manage your OVHcloud Connect services
* [ovhcloud pack-xdsl](ovhcloud_pack-xdsl.md)	 - Retrieve information and manage your PackXDSL services
* [ovhcloud sms](ovhcloud_sms.md)	 - Retrieve information and manage your SMS services
* [ovhcloud ssl](ovhcloud_ssl.md)	 - Retrieve information and manage your SSL services
* [ovhcloud ssl-gateway](ovhcloud_ssl-gateway.md)	 - Retrieve information and manage your SSL Gateway services
* [ovhcloud storage-netapp](ovhcloud_storage-netapp.md)	 - Retrieve information and manage your Storage NetApp services
* [ovhcloud support-tickets](ovhcloud_support-tickets.md)	 - Retrieve information and manage your support tickets
* [ovhcloud telephony](ovhcloud_telephony.md)	 - Retrieve information and manage your Telephony services
* [ovhcloud veeamcloudconnect](ovhcloud_veeamcloudconnect.md)	 - Retrieve information and manage your VeeamCloudConnect services
* [ovhcloud veeamenterprise](ovhcloud_veeamenterprise.md)	 - Retrieve information and manage your VeeamEnterprise services
* [ovhcloud version](ovhcloud_version.md)	 - Get OVHcloud CLI version
* [ovhcloud vmwareclouddirector-backup](ovhcloud_vmwareclouddirector-backup.md)	 - Retrieve information and manage your VMware Cloud Director Backup services
* [ovhcloud vmwareclouddirector-organization](ovhcloud_vmwareclouddirector-organization.md)	 - Retrieve information and manage your VMware Cloud Director Organizations
* [ovhcloud vps](ovhcloud_vps.md)	 - Retrieve information and manage your VPS services
* [ovhcloud vrack](ovhcloud_vrack.md)	 - Retrieve information and manage your vRack services
* [ovhcloud vrackservices](ovhcloud_vrackservices.md)	 - Retrieve information and manage your vRackServices services
* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud xdsl](ovhcloud_xdsl.md)	 - Retrieve information and manage your XDSL services

> **Tip**  Use `--json`, `--yaml`, or `--format` with a gval expression to integrate `ovhcloud` into scripts and automation pipelines.

---

## Examples

| Task                                  | Command                                        |
| ------------------------------------- | ---------------------------------------------- |
| Log in and save credentials           | `ovhcloud login`                                |
| List VPS instances (tabular)          | `ovhcloud vps list`                             |
| Fetch details of a single VPS in JSON | `ovhcloud vps get <service_id> --json`          |
| Reinstall a baremetal interactively   | `ovhcloud baremetal reinstall <id> --editor`    |

---

## Troubleshooting

* **Verbose output** — Use `--debug` to inspect raw API calls and responses.
* **Authentication issues** — Run `ovhcloud login` again to regenerate valid API keys.
* **Rate limits** — OVHcloud APIs impose rate limits; plan retries or exponential backoff in scripts.

---

## Further Reading

* OVHcloud API reference: [https://eu.api.ovh.com/console](https://eu.api.ovh.com/console)
* OVHcloud community guides and tutorials.
