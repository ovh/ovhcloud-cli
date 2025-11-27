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

* [ovhcloud webhosting module](ovhcloud_webhosting_module.md)	 - Manage one-click modules
