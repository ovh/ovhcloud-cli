## ovhcloud baremetal backup-agent create

Provision a Veeam backup agent for this server

### Synopsis

Provision a Veeam backup agent for this server.

The name, the addresses and the region are derived from the server itself, so nothing but the server name is needed. The agent is created NOT_INSTALLED: it protects nothing until the agent software runs on the machine, and retains nothing until it is put on a policy with `backup-agent edit --policy`.

```
ovhcloud baremetal backup-agent create <service_name> [flags]
```

### Options

```
      --display-name string   Name of the agent (default: agent-<service_name>)
      --dry-run               Print the call that would be made without making it
  -h, --help                  help for create
      --ip stringArray        Address the agent is reached at (default: the server's own address, in a /32)
      --region string         Region the agent operates in (default: the server's region)
      --wait                  Wait until the agent has settled before exiting
  -y, --yes                   Skip the confirmation prompt (required for unattended runs)
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
      --tenant string    Backup tenant to work on (default: the only one on the account)
      --vspc string      VSPC tenant to work on (default: the only one in the backup tenant)
```

### SEE ALSO

* [ovhcloud baremetal backup-agent](ovhcloud_baremetal_backup-agent.md)	 - Manage the Veeam backup agent protecting this server

