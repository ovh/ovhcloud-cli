## ovhcloud ip service confirm-termination

Confirm the termination of an IP service with the emailed token

```
ovhcloud ip service confirm-termination <service_name> <token> [flags]
```

### Options

```
      --commentary string   Free-text comment attached to the termination request
      --dry-run             Print the call that would be made without making it
      --future-use string   What comes next after this termination (press <tab> for the accepted values)
  -h, --help                help for confirm-termination
      --reason string       Why the service is being terminated (press <tab> for the accepted values)
  -y, --yes                 Skip the confirmation prompt (required for unattended runs)
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

* [ovhcloud ip service](ovhcloud_ip_service.md)	 - Manage your IP services: contacts, renewal and termination

