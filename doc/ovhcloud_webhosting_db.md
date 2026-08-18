## ovhcloud webhosting db

Manage databases

### Options

```
  -h, --help   help for db
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

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting db available-type](ovhcloud_webhosting_db_available-type.md)	 - List database types available for creation
* [ovhcloud webhosting db available-version](ovhcloud_webhosting_db_available-version.md)	 - List available versions for a database type
* [ovhcloud webhosting db capabilities](ovhcloud_webhosting_db_capabilities.md)	 - Get database capabilities
* [ovhcloud webhosting db change-password](ovhcloud_webhosting_db_change-password.md)	 - Change database password
* [ovhcloud webhosting db copy](ovhcloud_webhosting_db_copy.md)	 - Manage database copies
* [ovhcloud webhosting db create](ovhcloud_webhosting_db_create.md)	 - Create a database
* [ovhcloud webhosting db creation-capabilities](ovhcloud_webhosting_db_creation-capabilities.md)	 - List database creation capabilities
* [ovhcloud webhosting db delete](ovhcloud_webhosting_db_delete.md)	 - Delete a database
* [ovhcloud webhosting db dump](ovhcloud_webhosting_db_dump.md)	 - Manage database dumps
* [ovhcloud webhosting db get](ovhcloud_webhosting_db_get.md)	 - Get a database
* [ovhcloud webhosting db import](ovhcloud_webhosting_db_import.md)	 - Import a database dump
* [ovhcloud webhosting db list](ovhcloud_webhosting_db_list.md)	 - List databases
* [ovhcloud webhosting db request-action](ovhcloud_webhosting_db_request-action.md)	 - Request an action on a database
* [ovhcloud webhosting db restore](ovhcloud_webhosting_db_restore.md)	 - Restore database from snapshot date
* [ovhcloud webhosting db stats](ovhcloud_webhosting_db_stats.md)	 - Get database statistics

