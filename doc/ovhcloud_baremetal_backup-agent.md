## ovhcloud baremetal backup-agent

Manage the Veeam backup agent protecting this server

### Options

```
  -h, --help            help for backup-agent
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

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services
* [ovhcloud baremetal backup-agent create](ovhcloud_baremetal_backup-agent_create.md)	 - Provision a Veeam backup agent for this server
* [ovhcloud baremetal backup-agent delete](ovhcloud_baremetal_backup-agent_delete.md)	 - Remove the backup agent of this server — its restore points go with it
* [ovhcloud baremetal backup-agent edit](ovhcloud_baremetal_backup-agent_edit.md)	 - Change the backup agent of this server
* [ovhcloud baremetal backup-agent show](ovhcloud_baremetal_backup-agent_show.md)	 - Show the backup agent protecting this server

