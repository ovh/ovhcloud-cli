## ovhcloud cloud container-registry users

Manage container registry users

### Options

```
  -h, --help   help for users
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
```

### SEE ALSO

* [ovhcloud cloud container-registry](ovhcloud_cloud_container-registry.md)	 - Manage container registries in the given cloud project
* [ovhcloud cloud container-registry users create](ovhcloud_cloud_container-registry_users_create.md)	 - Create a new container registry user
* [ovhcloud cloud container-registry users delete](ovhcloud_cloud_container-registry_users_delete.md)	 - Delete a specific container registry user
* [ovhcloud cloud container-registry users get](ovhcloud_cloud_container-registry_users_get.md)	 - Get a specific container registry user
* [ovhcloud cloud container-registry users list](ovhcloud_cloud_container-registry_users_list.md)	 - List your container registry users
* [ovhcloud cloud container-registry users set-as-admin](ovhcloud_cloud_container-registry_users_set-as-admin.md)	 - Set a specific container registry user as admin

