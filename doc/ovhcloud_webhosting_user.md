## ovhcloud webhosting user

Create and manage the FTP/SSH users allowed to access your web hosting space.

```
  ovhcloud webhosting user [command]
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

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting user change-password](ovhcloud_webhosting_user_change-password.md)	 - Change FTP/SSH user password
* [ovhcloud webhosting user create](ovhcloud_webhosting_user_create.md)	 - Create a FTP/SSH user
* [ovhcloud webhosting user delete](ovhcloud_webhosting_user_delete.md)	 - Delete a FTP/SSH user
* [ovhcloud webhosting user get](ovhcloud_webhosting_user_get.md)	 - Get a FTP/SSH user
* [ovhcloud webhosting user list](ovhcloud_webhosting_user_list.md)	 - List FTP/SSH users
* [ovhcloud webhosting user update](ovhcloud_webhosting_user_update.md)	 - Update a FTP/SSH user
