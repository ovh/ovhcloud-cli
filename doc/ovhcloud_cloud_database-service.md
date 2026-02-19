## ovhcloud cloud database-service

Manage database services in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for database-service
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud database-service create](ovhcloud_cloud_database-service_create.md)	 - Create a new database service
* [ovhcloud cloud database-service database](ovhcloud_cloud_database-service_database.md)	 - Manage databases in a specific database service
* [ovhcloud cloud database-service delete](ovhcloud_cloud_database-service_delete.md)	 - Delete a specific database service
* [ovhcloud cloud database-service edit](ovhcloud_cloud_database-service_edit.md)	 - Edit a specific database service
* [ovhcloud cloud database-service get](ovhcloud_cloud_database-service_get.md)	 - Get a specific database service
* [ovhcloud cloud database-service list](ovhcloud_cloud_database-service_list.md)	 - List your database services

