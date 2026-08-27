## ovhcloud cloud managed-database user

Manage users in a specific managed database service

### Options

```
  -h, --help   help for user
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-database](ovhcloud_cloud_managed-database.md)	 - Manage managed database services in the given cloud project
* [ovhcloud cloud managed-database user create](ovhcloud_cloud_managed-database_user_create.md)	 - Create a new user in the given managed database service
* [ovhcloud cloud managed-database user credentials-reset](ovhcloud_cloud_managed-database_user_credentials-reset.md)	 - Reset the credentials of a specific user in the given managed database service
* [ovhcloud cloud managed-database user delete](ovhcloud_cloud_managed-database_user_delete.md)	 - Delete a specific user in the given managed database service
* [ovhcloud cloud managed-database user edit](ovhcloud_cloud_managed-database_user_edit.md)	 - Edit a user in the given managed database service
* [ovhcloud cloud managed-database user get](ovhcloud_cloud_managed-database_user_get.md)	 - Get a specific user in the given managed database service
* [ovhcloud cloud managed-database user list](ovhcloud_cloud_managed-database_user_list.md)	 - List all users in the given managed database service

