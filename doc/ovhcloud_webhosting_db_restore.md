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
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases
