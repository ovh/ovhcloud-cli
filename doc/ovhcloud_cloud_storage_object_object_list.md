## ovhcloud cloud storage object object list

List objects in the given storage container

```
ovhcloud cloud storage object object list <container_name> [flags]
```

### Options

```
  -h, --help                       help for list
      --key-marker string          Key marker for pagination
      --limit int                  Maximum number of objects to return (default 1000)
      --prefix string              Prefix to filter objects by name
      --version-id-marker string   Version ID marker for pagination
      --with-versions              Include object versions in the listing
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

