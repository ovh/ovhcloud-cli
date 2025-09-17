# OVHcloud official CLI

`ovhcloud` is a single, unified command‑line interface for managing the full range of OVHcloud products and account resources directly from your terminal. Whether you need to automate provisioning, perform quick look‑ups, or integrate OVHcloud operations into CI/CD pipelines, `ovhcloud` offers fine‑grained commands and consistent output formats (table, JSON, YAML, or custom gval expressions).

# Table of Contents

- [Installation](#installation)
- [Usage](#usage)
    - [Authenticating the CLI](#authenticating-the-cli)
    - [Examples](#examples)
- [Generate Shell Completion](#generate-shell-completion)
- [Contributing](#contributing)
    - [Build](#build)
    - [Run the tests](#run-the-tests)
    - [Our awesome contributors](#our-awesome-contributors)
- [Related links](#related-links)

# Installation

1. Download [latest release](https://github.com/ovh/ovhcloud-cli/releases/latest)
2. Untar / unzip the archive
3. Add the containing folder to your `PATH` environment variable

# Usage

```bash
$ ovhcloud [command] {subcommands} {parameters/flags}
```

Checkout the [full documentation](./doc/ovhcloud.md).

Available commands:
```
  account                          Manage your account
  alldom                           Retrieve information and manage your AllDom services
  baremetal                        Retrieve information and manage your baremetal services
  cdn-dedicated                    Retrieve information and manage your dedicated CDN services
  cloud                            Manage your projects and services in the Public Cloud universe
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
  ip                               Retrieve information and manage your Ip services
  iploadbalancing                  Retrieve information and manage your IpLoadbalancing services
  ldp                              Retrieve information and manage your Ldp services
  location                         Retrieve information and manage your Location services
  login                            Login to your OVHcloud account to create API credentials
  nutanix                          Retrieve information and manage your Nutanix services
  okms                             Retrieve information and manage your OKMS services
  overthebox                       Retrieve information and manage your OverTheBox services
  ovhcloudconnect                  Retrieve information and manage your OvhCloudConnect services
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
  vmwareclouddirector-backup       Retrieve information and manage your VmwareCloudDirectorBackup services
  vmwareclouddirector-organization Retrieve information and manage your VmwareCloudDirector Organizations
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

```bash
OVH_ENDPOINT=ovh-eu
OVH_APPLICATION_KEY=xxx
OVH_APPLICATION_SECRET=xxx
OVH_CONSUMER_KEY=xxx
OVH_CLOUD_PROJECT_SERVICE=<public cloud project ID> 
```

* Interactive login:
```bash
# Log in and create API credentials (interactive)
ovhcloud login
```

## Examples

| Task                                  | Command                                         |
| ------------------------------------- | ----------------------------------------------- |
| Log in and save credentials           | `ovhcloud login`                                |
| List VPS instances (tabular)          | `ovhcloud vps list`                             |
| Fetch details of a single VPS in JSON | `ovhcloud vps get <service_id> --json`          |
| Reinstall a baremetal interactively   | `ovhcloud baremetal reinstall <id> --editor`    |

# Generate Shell Completion

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

# Contributing

You've developed a new cool feature? Fixed an annoying bug? We'd be happy to hear from you, there are no small contributions!
 
Have a look in [CONTRIBUTING.md](https://github.com/ovh/ovhcloud-cli/blob/master/CONTRIBUTING.md)

## Build

```bash
# Build the OVHcloud cli
make build

# Cross-compile for other targets in ./dist
make release-snapshot

# Optionally, you can compile a WASM binary
make wasm
```

## Run the tests

```bash
make test
```

## Our awesome contributors

<a href="https://github.com/ovh/ovhcloud-cli/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=ovh/ovhcloud-cli" />
</a>

# Related links
 
 * Report bugs: https://github.com/ovh/ovhcloud-cli/issues
 * Get latest version: https://github.com/ovh/ovhcloud-cli/releases/latest
