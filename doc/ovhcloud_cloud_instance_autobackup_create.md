## ovhcloud cloud instance autobackup create

Create an automatic backup workflow for the given instance

```
ovhcloud cloud instance autobackup create <instance_id> [flags]
```

### Options

```
      --cron string    Unix Cron pattern (e.g. '0 0 * * *')
  -h, --help           help for create
      --name string    Workflow name
      --rotation int   Number of backups to keep
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud instance autobackup](ovhcloud_cloud_instance_autobackup.md)	 - Manage automatic backup workflows for instances

