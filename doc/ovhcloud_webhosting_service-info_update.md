## ovhcloud webhosting service-info update

Update service information

```
ovhcloud webhosting service-info update <service_name> [flags]
```

### Options

```
      --editor                       Use a text editor to define parameters
      --from-file string             File containing parameters
  -h, --help                         help for update
      --renew-automatic              Enable automatic renewal
      --renew-delete-at-expiration   Delete service at expiration
      --renew-forced                 Force renewal
      --renew-manual-payment         Enable manual payment for renewal
      --renew-period int             Renewal period in months
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
      --raw              Output the extracted value without JSON quoting (use with -o '<field>'), useful for scripting
                         Example:
                           --output 'id' --raw   (prints the id without surrounding quotes)
```

### SEE ALSO

* [ovhcloud webhosting service-info](ovhcloud_webhosting_service-info.md)	 - Manage webhosting service info

