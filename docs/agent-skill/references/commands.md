# Reference: install, auth & command map

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/ovh/ovhcloud-cli/main/install.sh | sh
# or: brew install ovh/ovhcloud-cli/ovhcloud
```

## Authenticate

```bash
ovhcloud login   # pick a region, paste an API token
```
Credentials live in an `ovh.conf` file (`/etc/ovh.conf`, `~/.ovh.conf`, or
`./ovh.conf`, in that priority) or in environment variables (`OVH_ENDPOINT`,
`OVH_APPLICATION_KEY`, `OVH_APPLICATION_SECRET`, `OVH_CONSUMER_KEY`). Create a
token at https://api.ovh.com/createToken/ (grant `GET/POST/PUT/DELETE` on `/*`
for full access). Multiple accounts: `ovhcloud login --profile <name>` then
`ovhcloud --profile <name> …`.

**Public Cloud** (`ovhcloud cloud …`) needs a project id: pass
`--cloud-project <id>` or set `OVH_CLOUD_PROJECT_SERVICE` (or `OS_TENANT_ID`).
Find it with `ovhcloud cloud project list`.

## Verbs (consistent across resources)

`list`/`ls`, `get <id>`, `create`, `edit <id>`, `delete <id>`. Nested resources
take their parents first (e.g. `… subnet get <networkId> <subnetId>`). Async
create/actions accept `--wait`.

## Discover any command

```bash
ovhcloud --help                       # all universes
ovhcloud <universe> --help            # its subcommands
ovhcloud <universe> <resource> --help # verbs + flags
ovhcloud <...> create --help          # flags + examples for one command
```
`create`/`edit` can also take a body from a file/editor: `--init-file ./p.json`
(write an example), `--from-file ./p.json`, `--editor`. CLI flags override them.

## Universe map (top-level groups)

- **Public Cloud** (`cloud`): `instance`, `network` (`private vrack`, `gateway`,
  `public`), `ip`, `loadbalancer`, `storage` (`object`, `block`, `file`),
  `managed-database`, `managed-kubernetes`, `managed-registry`, `managed-rancher`,
  `ai`, `ssh-key`, `user`, `region`, `quota`, `project`, `savings-plan`,
  `operation`.
- **Network**: `vrack`, `vrackservices`, `iploadbalancing`, `ip`,
  `ovhcloudconnect`, `ssl-gateway`, `cdn-dedicated`.
- **Domains & web**: `domain-name`, `domain-zone`, `webhosting`,
  `hosting-private-database`, `email-domain`, `email-pro`, `email-mxplan`, `ssl`.
- **Bare metal & private cloud**: `baremetal`, `dedicated-cloud`,
  `dedicated-ceph`, `dedicated-nasha`, `storage-netapp`, `nutanix`,
  `vmwareclouddirector-organization`, `vmwareclouddirector-backup`,
  `veeamcloudconnect`, `veeamenterprise`, `okms`.
- **Connectivity / telco**: `telephony`, `sms`, `xdsl`, `pack-xdsl`,
  `overthebox`, `ldp`.
- **Account & tooling**: `account` (`ssh-key`, `api`, `oauth2`), `iam`,
  `support-tickets`, `location`, `config`, `login`, `upgrade`, `version`,
  `browser` (experimental TUI), `completion`.

Run `ovhcloud <universe> --help` for the exhaustive, always-current list.
