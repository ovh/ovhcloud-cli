# OVHcloud official CLI

[![License](https://img.shields.io/github/license/ovh/ovhcloud-cli)](LICENSE)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/ovh/ovhcloud-cli)](https://github.com/ovh/ovhcloud-cli/releases/latest)
[![Homebrew](https://img.shields.io/badge/dynamic/regex?url=https%3A%2F%2Fraw.githubusercontent.com%2Fovh%2Fhomebrew-tap%2Frefs%2Fheads%2Fmain%2FCasks%2Fovhcloud-cli.rb&search=version%20%22%28%3F%3Cversion%3E%5B%5E%22%5D%2B%29%22&replace=%24%3Cversion%3E&label=homebrew)](#install-using-homebrew)
[![Release Date](https://img.shields.io/github/release-date/ovh/ovhcloud-cli)](https://github.com/ovh/ovhcloud-cli/releases)
[![Download Count](https://img.shields.io/github/downloads/ovh/ovhcloud-cli/total.svg)](https://github.com/ovh/ovhcloud-cli/releases/latest)
[![Discord](https://img.shields.io/badge/Discord-ovhcloud-blue.svg)](https://discord.gg/ovhcloud)

[![Packaging status](https://repology.org/badge/vertical-allrepos/ovhcloud-cli.svg)](https://repology.org/project/ovhcloud-cli/versions)

`ovhcloud` is a single, unified command‑line interface for managing the full range of OVHcloud products and account resources directly from your terminal. Whether you need to automate provisioning, perform quick look‑ups, or integrate OVHcloud operations into CI/CD pipelines, `ovhcloud` offers fine‑grained commands and consistent output formats (table, JSON, YAML, or custom gval expressions).

# Table of Contents

- [Installation](#installation)
    - [Installation command](#installation-command)
    - [Binary download](#binary-download)
    - [Run with Docker](#run-with-docker)
    - [Install using HomeBrew](#install-using-homebrew)
    - [Install from the source](#install-from-the-source)
- [Usage](#usage)
    - [Authenticating the CLI](#authenticating-the-cli)
    - [Examples](#examples)
- [Available products](#available-products)
- [Generate Shell Completion](#generate-shell-completion)
- [Contributing](#contributing)
    - [Build](#build)
    - [Run the tests](#run-the-tests)
    - [Our awesome contributors](#our-awesome-contributors)
- [Related links](#related-links)

# Installation

## Installation command

To install the OVHcloud CLI, you can use the following command:

```sh
curl -fsSL https://raw.githubusercontent.com/ovh/ovhcloud-cli/main/install.sh | sh
```

## Binary download

1. Download [latest release](https://github.com/ovh/ovhcloud-cli/releases/latest)
2. Untar / unzip the archive
3. Add the containing folder to your `PATH` environment variable

## Run with Docker

You can also run the CLI using Docker:

```sh
docker run -it --rm -v ovhcloud-cli-config-files:/config ovhcom/ovhcloud-cli login
```

## Install using Homebrew

```sh
brew install --cask ovh/tap/ovhcloud-cli
```

## Install from the source

Requires Go to be installed on your system.

```sh
go install github.com/ovh/ovhcloud-cli/cmd/ovhcloud@latest
```

# Usage

```sh
$ ovhcloud [command] {subcommands} {parameters/flags}
```

Checkout the [full documentation](./doc/ovhcloud.md).

Available commands:
```
  account                          Manage your account
  alldom                           Retrieve information and manage your AllDom services
  baremetal                        Retrieve information and manage your Bare Metal services
  cdn-dedicated                    Retrieve information and manage your dedicated CDN services
  cloud                            Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
  completion                       Generate the autocompletion script for the specified shell
  config                           Manage your CLI configuration
  dedicated-ceph                   Retrieve information and manage your Dedicated Ceph services
  dedicated-cloud                  Retrieve information and manage your DedicatedCloud services
  dedicated-cluster                Retrieve information and manage your DedicatedCluster services
  dedicated-nasha                  Retrieve information and manage your Dedicated NasHA services
  domain-name                      Retrieve information and manage your domain names
  domain-zone                      Retrieve information and manage your domain zones
  email-domain                     Retrieve information and manage your Email Domain services
  email-mxplan                     Retrieve information and manage your Email MXPlan services
  email-pro                        Retrieve information and manage your EmailPro services
  help                             Help about any command
  hosting-private-database         Retrieve information and manage your HostingPrivateDatabase services
  iam                              Manage IAM resources, permissions and policies
  ip                               Retrieve information and manage your IP services
  iploadbalancing                  Retrieve information and manage your IP LoadBalancing services
  ldp                              Retrieve information and manage your LDP (Logs Data Platform) services
  location                         Retrieve information and manage your Location services
  login                            Login to your OVHcloud account to create API credentials
  nutanix                          Retrieve information and manage your Nutanix services
  okms                             Retrieve information and manage your OKMS services
  overthebox                       Retrieve information and manage your OverTheBox services
  ovhcloudconnect                  Retrieve information and manage your OVHcloud Connect services
  pack-xdsl                        Retrieve information and manage your PackXDSL services
  sms                              Retrieve information and manage your SMS services
  ssl                              Retrieve information and manage your SSL services
  ssl-gateway                      Retrieve information and manage your SSL Gateway services
  storage-netapp                   Retrieve information and manage your Storage NetApp services
  support-tickets                  Retrieve information and manage your support tickets
  telephony                        Retrieve information and manage your Telephony services
  veeamcloudconnect                Retrieve information and manage your VeeamCloudConnect services
  veeamenterprise                  Retrieve information and manage your VeeamEnterprise services
  version                          Get OVHcloud CLI version
  vmwareclouddirector-backup       Retrieve information and manage your VMware Cloud Director Backup services
  vmwareclouddirector-organization Retrieve information and manage your VMware Cloud Director Organizations
  vps                              Retrieve information and manage your VPS services
  vrack                            Retrieve information and manage your vRack services
  vrackservices                    Retrieve information and manage your vRackServices services
  webhosting                       Retrieve information and manage your WebHosting services
  xdsl                             Retrieve information and manage your XDSL services
```

Global options:

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using gval format)
  -h, --help            help for ovhcloud
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

## Authenticating the CLI

OVHcloud CLI requires authentication to be able to make API calls. There are several ways to define your credentials.

Check out the [authentication page](./doc/authentication.md) for further information about the configuration and the authentication means.

* Interactive login:
```sh
# Log in and create API credentials (interactive)
ovhcloud login
```

* Using a configuration file:

Default settings can be set using a configuration file named `.ovh.conf` and located in your `${HOME}` directory.

Example of configuration file:

```ini
[default]
endpoint = ovh-eu

[ovh-eu]
client_id     = <OAuth 2.0 client ID>
client_secret = <OAuth 2.0 client secret>

[ovh-cli]
default_cloud_project = <public cloud project ID>
```

* Using environment variables:

```sh
OVH_ENDPOINT=ovh-eu
OVH_APPLICATION_KEY=xxx
OVH_APPLICATION_SECRET=xxx
OVH_CONSUMER_KEY=xxx
OVH_CLOUD_PROJECT_SERVICE=<public cloud project ID> 
```

## Examples

| Task                                     | Command                                         |
| ---------------------------------------- | ----------------------------------------------- |
| Log in and save credentials              | `ovhcloud login`                                |
| List VPS instances (tabular)             | `ovhcloud vps list`                             |
| Fetch details of a single VPS in JSON    | `ovhcloud vps get <service_id> --json`          |
| Reinstall a baremetal interactively      | `ovhcloud baremetal reinstall <id> --editor`    |
| List instances and filter on GRA9 region | `ovhcloud cloud instance list --filter 'region=="GRA9"'`|

# Available products

| Category                              | Command                         | Covered    |
|---------------------------------------|---------------------------------|------------|
| **Backup**                            | veeamcloudconnect               | Partially  |
|                                       | veeamenterprise                 | Partially  |
| **Communication**                     | sms                             | Partially  |
|                                       | telephony                       | Partially  |
| **Compute / Dedicated / VPS**         | baremetal                       | Yes        |
|                                       | vps                             | Yes        |
| **Connectivity**                      | pack-xdsl                       | Partially  |
|                                       | xdsl                            | Partially  |
| **Domains / DNS**                     | alldom                          | Partially  |
|                                       | domain-name                     | Partially  |
|                                       | domain-zone                     | Partially  |
| **Email**                             | email-domain                    | Partially  |
|                                       | email-mxplan                    | Partially  |
|                                       | email-pro                       | Partially  |
| **Hybrid Cloud**                      | nutanix                         | Partially  |
| **Hosting**                           | webhosting                      | Partially  |
|                                       | hosting-private-database        | Partially  |
| **Identity / Account / Access**       | account                         | Partially  |
|                                       | iam                             | Yes        |
|                                       | login                           | Partially  |
| **Infra Meta**                        | location                        | Partially  |
| **Network**                           | ip                              | Partially  |
|                                       | overthebox                      | Partially  |
|                                       | vrack                           | Partially  |
|                                       | vrackservices                   | Partially  |
| **Network / Acceleration**            | cdn-dedicated                   | Partially  |
| **Network / Connectivity**            | ovhcloudconnect                 | Partially  |
| **Network / Load Balancing**          | iploadbalancing                 | Partially  |
| **Observability**                     | ldp                             | Partially  |
| **Private Cloud**                     | dedicated-cloud                 | Partially  |
|                                       | dedicated-cluster               | Partially  |
| **Public Cloud / Access**             | cloud ssh-key                   | Yes        |
| **Public Cloud / Compute**            | cloud instance                  | Yes        |
| **Public Cloud / Container Registry** | cloud container-registry        | Yes        |
| **Public Cloud / Containers**         | cloud kube                      | Yes        |
| **Public Cloud / Orchestration**      | cloud rancher                   | Yes        |
| **Public Cloud / Databases**          | cloud database-service          | Partially  |
| **Public Cloud / Governance**         | cloud quota                     | Yes        |
| **Public Cloud / Meta**               | cloud reference                 | Yes        |
|                                       | cloud region                    | Yes        |
| **Public Cloud / Network**            | cloud loadbalancer              | Partially  |
|                                       | cloud network                   | Yes        |
| **Public Cloud / Object Storage**     | cloud storage-s3                | Yes        |
|                                       | cloud storage-swift             | Yes        |
| **Public Cloud / Ops**                | cloud operation                 | Yes        |
| **Public Cloud / Project**            | cloud project                   | Yes        |
| **Public Cloud / Storage**            | cloud storage-block             | Yes        |
| **Public Cloud / Identity**           | cloud user                      | Yes        |
| **Security**                          | ssl                             | Partially  |
|                                       | okms                            | Partially  |
| **Security / Edge**                   | ssl-gateway                     | Partially  |
| **Storage**                           | dedicated-ceph                  | Partially  |
|                                       | dedicated-nasha                 | Partially  |
|                                       | storage-netapp                  | Partially  |
| **Support**                           | support-tickets                 | Partially  |
| **VMware**                            | vmwareclouddirector-organization| Partially  |
|                                       | vmwareclouddirector-backup      | Partially  |
|---------------------------------------|---------------------------------|------------|

# Generate Shell Completion

```sh
# Bash
eval "$(ovhcloud completion bash)"
# Zsh
eval "$(ovhcloud completion zsh)"
# Fish
ovhcloud completion fish | source
# PowerShell
ovhcloud completion powershell | Out-String | Invoke-Expression
```

Add the appropriate line to your shell’s startup file (`~/.bashrc`, `~/.zshrc`, etc.) to enable persistent autocompletion.

# Contributing

You've developed a new cool feature? Fixed an annoying bug? We'd be happy to hear from you, there are no small contributions!
 
Have a look in [CONTRIBUTING.md](https://github.com/ovh/ovhcloud-cli/blob/main/CONTRIBUTING.md)

## Build

```sh
# Build the OVHcloud cli
make all

# Cross-compile for other targets in ./dist
make release-snapshot

# Optionally, you can compile a WASM binary
make wasm
```

## Run the tests

```sh
make test
```

## Our awesome contributors

<a href="https://github.com/ovh/ovhcloud-cli/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=ovh/ovhcloud-cli" />
</a>

# Related links
 
 * Report bugs: https://github.com/ovh/ovhcloud-cli/issues
 * Get latest version: https://github.com/ovh/ovhcloud-cli/releases/latest
