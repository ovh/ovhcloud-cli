## ovhcloud webhosting extra-sql service-info

Manage extra SQL service info

```
  ovhcloud webhosting extra-sql service-info [command]
```

### Options

```
  -h, --help   help for service-info
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

* [ovhcloud webhosting extra-sql](ovhcloud_webhosting_extra-sql.md)	 - Manage extra SQL options
* [ovhcloud webhosting extra-sql service-info get](ovhcloud_webhosting_extra-sql_service-info_get.md)	 - Get extra SQL service information
* [ovhcloud webhosting extra-sql service-info update](ovhcloud_webhosting_extra-sql_service-info_update.md)	 - Update extra SQL service information
