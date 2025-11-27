## ovhcloud webhosting cdn operation

Manage CDN operations

```
  ovhcloud webhosting cdn operation [command]
```

### Options

```
  -h, --help   help for operation
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

* [ovhcloud webhosting cdn](ovhcloud_webhosting_cdn.md)	 - Manage CDN
* [ovhcloud webhosting cdn operation get](ovhcloud_webhosting_cdn_operation_get.md)	 - Get a CDN operation
* [ovhcloud webhosting cdn operation list](ovhcloud_webhosting_cdn_operation_list.md)	 - List CDN operations
