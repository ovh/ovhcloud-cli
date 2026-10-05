## ovhcloud baremetal service-info edit

Edit service information of the given baremetal

```
ovhcloud baremetal service-info edit <service_name> [flags]
```

### Options

```
      --editor                       Use a text editor to define parameters
  -h, --help                         help for edit
      --renew-automatic              Renew the service automatically
      --renew-delete-at-expiration   Delete the service when it expires
      --renew-forced                 Force the renewal
      --renew-manual-payment         Pay the renewal manually
      --renew-period int             Renewal period, in months
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

* [ovhcloud baremetal service-info](ovhcloud_baremetal_service-info.md)	 - Manage service information of the given baremetal

