## ovhcloud cloud storage-s3 object edit

Edit the given object in the storage container

```
ovhcloud cloud storage-s3 object edit <container_name> <object_name> [flags]
```

### Options

```
      --editor                     Use a text editor to define parameters
  -h, --help                       help for edit
      --legal-hold string          Legal hold status (on, off)
      --lock-mode string           Lock mode (compliance, governance)
      --lock-retain-until string   Lock retain until date (e.g., 2024-12-31T23:59:59Z)
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

