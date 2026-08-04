## ovhcloud webhosting extra-sql

Manage extra SQL options

### Options

```
  -h, --help   help for extra-sql
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

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Manage Web Hosting (databases, cron, SSL, env vars, logs)
* [ovhcloud webhosting extra-sql databases](ovhcloud_webhosting_extra-sql_databases.md)	 - List databases linked to an extra SQL option
* [ovhcloud webhosting extra-sql get](ovhcloud_webhosting_extra-sql_get.md)	 - Get an extra SQL option
* [ovhcloud webhosting extra-sql list](ovhcloud_webhosting_extra-sql_list.md)	 - List extra SQL options
* [ovhcloud webhosting extra-sql service-info](ovhcloud_webhosting_extra-sql_service-info.md)	 - Manage extra SQL service info
* [ovhcloud webhosting extra-sql terminate](ovhcloud_webhosting_extra-sql_terminate.md)	 - Terminate an extra SQL option

