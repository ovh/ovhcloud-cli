## ovhcloud backup-services vspc get

Show one VSPC tenant

```
ovhcloud backup-services vspc get [<vspc_id>] [flags]
```

### Options

```
  -h, --help   help for get
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

* [ovhcloud backup-services vspc](ovhcloud_backup-services_vspc.md)	 - Show and rename the Veeam Service Provider Console tenants

