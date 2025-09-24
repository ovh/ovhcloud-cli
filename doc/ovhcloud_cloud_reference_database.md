## ovhcloud cloud reference database

Fetch database reference data in the given cloud project

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

* [ovhcloud cloud reference](ovhcloud_cloud_reference.md)	 - Fetch reference data in the given cloud project
* [ovhcloud cloud reference database list-engines](ovhcloud_cloud_reference_database_list-engines.md)	 - List available database engines in the given cloud project
* [ovhcloud cloud reference database list-node-flavors](ovhcloud_cloud_reference_database_list-node-flavors.md)	 - List available database node flavors in the given cloud project
* [ovhcloud cloud reference database list-plans](ovhcloud_cloud_reference_database_list-plans.md)	 - List available database plans in the given cloud project

