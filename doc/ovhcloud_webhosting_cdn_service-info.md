## ovhcloud webhosting cdn service-info

Manage CDN service info

### Options

```
  -h, --help   help for service-info
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
  -y, --yes              Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud webhosting cdn](ovhcloud_webhosting_cdn.md)	 - Manage CDN
* [ovhcloud webhosting cdn service-info get](ovhcloud_webhosting_cdn_service-info_get.md)	 - Get CDN service information
* [ovhcloud webhosting cdn service-info update](ovhcloud_webhosting_cdn_service-info_update.md)	 - Update CDN service information

