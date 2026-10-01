## ovhcloud backup-services

Retrieve information and manage your Veeam Backup services

### Synopsis

Retrieve information and manage your Veeam Backup services.

A backup tenant holds storage vaults and a Veeam Service Provider Console tenant; the console tenant is what drives the agents installed on your machines. Both levels are resolved when the account has only one of them, so --tenant and --vspc are only needed when there is a choice to make.

The agents themselves are managed from the machine they protect: see `ovhcloud baremetal backup-agent`.

### Options

```
  -h, --help            help for backup-services
      --tenant string   Backup tenant to work on (default: the only one on the account)
      --vspc string     VSPC tenant to work on (default: the only one in the backup tenant)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
                         Examples:
                           --output json
                           --output yaml
                           --output interactive
                           --output 'id' (to extract a single field)
                           --output 'nested.field.subfield' (to extract a nested field)
                           --output '[id, "name"]' (to extract multiple fields as an array)
                           --output '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                           --output 'name+","+type' (to extract and concatenate fields in a string)
                           --output '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud backup-services agents](ovhcloud_backup-services_agents.md)	 - List every backup agent, and what each one protects
* [ovhcloud backup-services billing](ovhcloud_backup-services_billing.md)	 - Show what each part of the backup service costs, and what it has consumed
* [ovhcloud backup-services deploy-script](ovhcloud_backup-services_deploy-script.md)	 - Show the command that installs the backup agent on a machine
* [ovhcloud backup-services licenses](ovhcloud_backup-services_licenses.md)	 - Show the Veeam licences held by a VSPC tenant
* [ovhcloud backup-services policies](ovhcloud_backup-services_policies.md)	 - List the retention policies an agent can be put on
* [ovhcloud backup-services tenant](ovhcloud_backup-services_tenant.md)	 - Show the backup tenants of the account
* [ovhcloud backup-services vault](ovhcloud_backup-services_vault.md)	 - Show and rename the storage vaults of a backup tenant
* [ovhcloud backup-services vspc](ovhcloud_backup-services_vspc.md)	 - Show and rename the Veeam Service Provider Console tenants

