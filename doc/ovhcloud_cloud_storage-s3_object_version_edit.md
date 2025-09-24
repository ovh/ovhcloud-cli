## ovhcloud cloud storage-s3 object version edit

Edit the given version of an object in the storage container

```
ovhcloud cloud storage-s3 object version edit <container_name> <object_name> <version_id> [flags]
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

* [ovhcloud cloud storage-s3 object version](ovhcloud_cloud_storage-s3_object_version.md)	 - Manage versions of objects in the given storage container

