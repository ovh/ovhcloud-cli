## ovhcloud webhosting own-log user

Manage own log users

```
  ovhcloud webhosting own-log user [command]
```

### Options

```
  -h, --help   help for user
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

* [ovhcloud webhosting own-log](ovhcloud_webhosting_own-log.md)	 - Manage own logs
* [ovhcloud webhosting own-log user change-password](ovhcloud_webhosting_own-log_user_change-password.md)	 - Change an own log user password
* [ovhcloud webhosting own-log user create](ovhcloud_webhosting_own-log_user_create.md)	 - Create an own log user
* [ovhcloud webhosting own-log user delete](ovhcloud_webhosting_own-log_user_delete.md)	 - Delete an own log user
* [ovhcloud webhosting own-log user get](ovhcloud_webhosting_own-log_user_get.md)	 - Get an own log user
* [ovhcloud webhosting own-log user list](ovhcloud_webhosting_own-log_user_list.md)	 - List users for an own log
* [ovhcloud webhosting own-log user update](ovhcloud_webhosting_own-log_user_update.md)	 - Update an own log user
