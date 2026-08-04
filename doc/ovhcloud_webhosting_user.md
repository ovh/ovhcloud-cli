## ovhcloud webhosting user

Manage FTP/SSH users

### Synopsis

Create and manage the FTP/SSH users allowed to access your web hosting space.

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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Manage Web Hosting (databases, cron, SSL, env vars, logs)
* [ovhcloud webhosting user change-password](ovhcloud_webhosting_user_change-password.md)	 - Change FTP/SSH user password
* [ovhcloud webhosting user create](ovhcloud_webhosting_user_create.md)	 - Create a FTP/SSH user
* [ovhcloud webhosting user delete](ovhcloud_webhosting_user_delete.md)	 - Delete a FTP/SSH user
* [ovhcloud webhosting user get](ovhcloud_webhosting_user_get.md)	 - Get a FTP/SSH user
* [ovhcloud webhosting user list](ovhcloud_webhosting_user_list.md)	 - List FTP/SSH users
* [ovhcloud webhosting user update](ovhcloud_webhosting_user_update.md)	 - Update a FTP/SSH user

