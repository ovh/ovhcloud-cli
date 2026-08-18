## ovhcloud baremetal power off

Power off the given baremetal

```
ovhcloud baremetal power off <service_name> [flags]
```

### Options

```
      --dry-run            Print the two calls that would be made without making them
  -h, --help               help for off
      --timeout duration   How long --wait waits (default 10m0s)
      --wait               Wait until the server reports itself off before exiting
  -y, --yes                Skip the confirmation prompt (required for unattended runs)
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

* [ovhcloud baremetal power](ovhcloud_baremetal_power.md)	 - Power the given baremetal off and on

