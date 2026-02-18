## ovhcloud cloud storage-s3 object copy

Copy the given object to another bucket or key

```
ovhcloud cloud storage-s3 object copy <container_name> <object_name> [flags]
```

### Options

```
  -h, --help                   help for copy
      --storage-class string   Target storage class (HIGH_PERF, STANDARD, STANDARD_IA)
      --target-bucket string   Target bucket name
      --target-key string      Target object key
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

* [ovhcloud cloud storage-s3 object](ovhcloud_cloud_storage-s3_object.md)	 - Manage objects in the given storage container

