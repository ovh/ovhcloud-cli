## ovhcloud webhosting cdn domain option

Manage CDN domain options

### Options

```
  -h, --help   help for option
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

* [ovhcloud webhosting cdn domain](ovhcloud_webhosting_cdn_domain.md)	 - Manage CDN domains
* [ovhcloud webhosting cdn domain option add](ovhcloud_webhosting_cdn_domain_option_add.md)	 - Add a CDN domain option
* [ovhcloud webhosting cdn domain option delete](ovhcloud_webhosting_cdn_domain_option_delete.md)	 - Delete a CDN domain option
* [ovhcloud webhosting cdn domain option get](ovhcloud_webhosting_cdn_domain_option_get.md)	 - Get CDN domain option details
* [ovhcloud webhosting cdn domain option list](ovhcloud_webhosting_cdn_domain_option_list.md)	 - List CDN domain options
* [ovhcloud webhosting cdn domain option update](ovhcloud_webhosting_cdn_domain_option_update.md)	 - Update a CDN domain option

