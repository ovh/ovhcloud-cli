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
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud database-service](ovhcloud_cloud_database-service.md)	 - Manage database services in the given cloud project
* [ovhcloud cloud database-service database create](ovhcloud_cloud_database-service_database_create.md)	 - Create a new database in the given database service
* [ovhcloud cloud database-service database delete](ovhcloud_cloud_database-service_database_delete.md)	 - Delete a specific database in the given database service
* [ovhcloud cloud database-service database get](ovhcloud_cloud_database-service_database_get.md)	 - Get a specific database in the given database service
* [ovhcloud cloud database-service database list](ovhcloud_cloud_database-service_database_list.md)	 - List all databases in the given database service

