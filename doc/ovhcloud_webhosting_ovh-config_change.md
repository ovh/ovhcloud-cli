## ovhcloud webhosting ovh-config change

Change a .ovhconfig entry

```
ovhcloud webhosting ovh-config change <service_name> <id> [flags]
```

### Options

```
      --container string        Container image
      --editor                  Use a text editor to define parameters
      --engine-name string      Engine name
      --engine-version string   Engine version
      --environment string      Environment (production, development, ...)
      --from-file string        File containing parameters
  -h, --help                    help for change
      --http-firewall string    HTTP firewall mode (none, security, ...)
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

* [ovhcloud webhosting ovh-config](ovhcloud_webhosting_ovh-config.md)	 - Manage .ovhconfig settings

