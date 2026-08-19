## ovhcloud ssl-gateway list

List your SSL Gateway services

```
ovhcloud ssl-gateway list [flags]
```

### Options

```
      --filter stringArray   Filter results by any property using https://github.com/PaesslerAG/gval syntax
                             Examples:
                               --filter 'state=="running"'
                               --filter 'name=~"^my.*"'
                               --filter 'nested.property.subproperty>10'
                               --filter 'startDate>="2023-12-01"'
                               --filter 'name=~"something" && nbField>10'
  -h, --help                 help for list
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud ssl-gateway](ovhcloud_ssl-gateway.md)	 - Retrieve information and manage your SSL Gateway services

