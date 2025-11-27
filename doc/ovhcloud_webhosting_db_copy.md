## ovhcloud webhosting db copy

Manage database copies

```
  ovhcloud webhosting db copy [command]
```

### Options

```
  -h, --help   help for copy
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

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases
* [ovhcloud webhosting db copy create](ovhcloud_webhosting_db_copy_create.md)	 - Create a database copy
* [ovhcloud webhosting db copy delete](ovhcloud_webhosting_db_copy_delete.md)	 - Delete a database copy
* [ovhcloud webhosting db copy get](ovhcloud_webhosting_db_copy_get.md)	 - Get a database copy
* [ovhcloud webhosting db copy list](ovhcloud_webhosting_db_copy_list.md)	 - List database copies
* [ovhcloud webhosting db copy restore](ovhcloud_webhosting_db_copy_restore.md)	 - Restore a database copy
