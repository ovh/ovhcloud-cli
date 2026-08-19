## ovhcloud cloud storage object object edit

Edit the given object in the storage container

```
ovhcloud cloud storage object object edit <container_name> <object_name> [flags]
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
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud storage object object](ovhcloud_cloud_storage_object_object.md)	 - Manage objects in the given storage container

