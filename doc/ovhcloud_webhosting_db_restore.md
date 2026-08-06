## ovhcloud webhosting db restore

Restore database from snapshot date

```
ovhcloud webhosting db restore <service_name> <name> [flags]
```

### Options

```
      --date string        Dump type to restore (allowed: daily.1, now, weekly.1)
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for restore
      --send-email         Send email when restore completes
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
  -y, --yes              Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases

