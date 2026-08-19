## ovhcloud cloud storage object object restore

Restore the given object from archival storage

```
ovhcloud cloud storage object object restore <container_name> <object_name> [flags]
```

### Options

```
      --days int   Number of days the restored object will be available
  -h, --help       help for restore
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

