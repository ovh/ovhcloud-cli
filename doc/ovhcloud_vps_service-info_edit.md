## ovhcloud vps service-info edit

Edit service information for the given VPS

```
ovhcloud vps service-info edit <service_name> [flags]
```

### Options

```
      --editor                       Use a text editor to define parameters
  -h, --help                         help for edit
      --renew-automatic              Enable automatic renewal
      --renew-delete-at-expiration   Delete service at expiration
      --renew-forced                 Force renewal
      --renew-manual-payment         Enable manual payment for renewal
      --renew-period int             Renewal period (in months)
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

* [ovhcloud vps service-info](ovhcloud_vps_service-info.md)	 - Manage service information for the given VPS

