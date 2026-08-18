## ovhcloud webhosting module install

Install a module

```
ovhcloud webhosting module install <service_name> [flags]
```

### Options

```
      --admin string            Admin login
      --admin-password string   Admin password
      --domain string           Domain
      --editor                  Use a text editor to define parameters
      --from-file string        File containing parameters
  -h, --help                    help for install
      --language string         Language
      --module-id int           Module ID
      --module-name string      Module name (latest version will be selected)
      --path string             Install path
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

* [ovhcloud webhosting module](ovhcloud_webhosting_module.md)	 - Manage one-click modules

