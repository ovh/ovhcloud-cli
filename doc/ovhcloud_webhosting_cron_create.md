## ovhcloud webhosting cron create

Create a cron task

```
  ovhcloud webhosting cron create <service_name> [flags]
```

### Options

```
      --command string       Command to execute
      --description string   Description
      --editor               Use a text editor to define parameters
      --email string         Email for stderr
      --frequency string     Frequency (crontab format)
      --from-file string     File containing parameters
  -h, --help                 help for create
      --language string      Language
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

* [ovhcloud webhosting cron](ovhcloud_webhosting_cron.md)	 - Manage cron tasks
