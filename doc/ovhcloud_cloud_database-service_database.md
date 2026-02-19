## ovhcloud cloud database-service database

Manage databases in a specific database service

### Options

```
  -h, --help   help for database
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

* [ovhcloud cloud database-service](ovhcloud_cloud_database-service.md)	 - Manage database services in the given cloud project
* [ovhcloud cloud database-service database create](ovhcloud_cloud_database-service_database_create.md)	 - Create a new database in the given database service
* [ovhcloud cloud database-service database delete](ovhcloud_cloud_database-service_database_delete.md)	 - Delete a specific database in the given database service
* [ovhcloud cloud database-service database get](ovhcloud_cloud_database-service_database_get.md)	 - Get a specific database in the given database service
* [ovhcloud cloud database-service database list](ovhcloud_cloud_database-service_database_list.md)	 - List all databases in the given database service

