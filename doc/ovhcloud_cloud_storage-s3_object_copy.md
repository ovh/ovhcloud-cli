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

* [ovhcloud cloud storage-s3 object](ovhcloud_cloud_storage-s3_object.md)	 - Manage objects in the given storage container

