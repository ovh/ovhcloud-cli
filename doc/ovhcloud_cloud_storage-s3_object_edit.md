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
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud storage-s3 object](ovhcloud_cloud_storage-s3_object.md)	 - Manage objects in the given storage container

