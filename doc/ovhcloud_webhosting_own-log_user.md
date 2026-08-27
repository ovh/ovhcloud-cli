## ovhcloud webhosting own-log user

Manage own log users

### Options

```
  -h, --help   help for user
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
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting own-log](ovhcloud_webhosting_own-log.md)	 - Manage own logs
* [ovhcloud webhosting own-log user change-password](ovhcloud_webhosting_own-log_user_change-password.md)	 - Change an own log user password
* [ovhcloud webhosting own-log user create](ovhcloud_webhosting_own-log_user_create.md)	 - Create an own log user
* [ovhcloud webhosting own-log user delete](ovhcloud_webhosting_own-log_user_delete.md)	 - Delete an own log user
* [ovhcloud webhosting own-log user get](ovhcloud_webhosting_own-log_user_get.md)	 - Get an own log user
* [ovhcloud webhosting own-log user list](ovhcloud_webhosting_own-log_user_list.md)	 - List users for an own log
* [ovhcloud webhosting own-log user update](ovhcloud_webhosting_own-log_user_update.md)	 - Update an own log user

