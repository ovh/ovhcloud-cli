## ovhcloud webhosting module catalog

Browse available one-click modules

### Options

```
  -h, --help   help for catalog
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
* [ovhcloud webhosting module catalog get](ovhcloud_webhosting_module_catalog_get.md)	 - Get available module details
* [ovhcloud webhosting module catalog list](ovhcloud_webhosting_module_catalog_list.md)	 - List available modules

