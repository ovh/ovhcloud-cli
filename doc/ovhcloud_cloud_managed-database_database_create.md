## ovhcloud cloud managed-database database create

Create a new database in the given managed database service

### Synopsis

Use this command to create a database in the given managed database service.

	ovhcloud cloud managed-database database create <service_id> --name mydb


```
ovhcloud cloud managed-database database create <service_id> [flags]
```

### Options

```
  -h, --help          help for create
      --name string   Name of the database to create
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

* [ovhcloud cloud managed-database database](ovhcloud_cloud_managed-database_database.md)	 - Manage databases in a specific managed database service

