## ovhcloud webhosting extra-sql

Manage extra SQL options

```
  ovhcloud webhosting extra-sql [command]
```

### Options

```
  -h, --help   help for extra-sql
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

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting extra-sql databases](ovhcloud_webhosting_extra-sql_databases.md)	 - List databases linked to an extra SQL option
* [ovhcloud webhosting extra-sql get](ovhcloud_webhosting_extra-sql_get.md)	 - Get an extra SQL option
* [ovhcloud webhosting extra-sql list](ovhcloud_webhosting_extra-sql_list.md)	 - List extra SQL options
* [ovhcloud webhosting extra-sql service-info](ovhcloud_webhosting_extra-sql_service-info.md)	 - Manage extra SQL service info
* [ovhcloud webhosting extra-sql terminate](ovhcloud_webhosting_extra-sql_terminate.md)	 - Terminate an extra SQL option
