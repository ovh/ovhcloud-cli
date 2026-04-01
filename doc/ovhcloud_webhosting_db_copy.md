## ovhcloud webhosting db copy

Manage database copies

### Options

```
  -h, --help   help for copy
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

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases
* [ovhcloud webhosting db copy create](ovhcloud_webhosting_db_copy_create.md)	 - Create a database copy
* [ovhcloud webhosting db copy delete](ovhcloud_webhosting_db_copy_delete.md)	 - Delete a database copy
* [ovhcloud webhosting db copy get](ovhcloud_webhosting_db_copy_get.md)	 - Get a database copy
* [ovhcloud webhosting db copy list](ovhcloud_webhosting_db_copy_list.md)	 - List database copies
* [ovhcloud webhosting db copy restore](ovhcloud_webhosting_db_copy_restore.md)	 - Restore a database copy

