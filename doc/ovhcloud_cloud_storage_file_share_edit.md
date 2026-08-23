## ovhcloud cloud storage file share edit

Edit the given share

```
ovhcloud cloud storage file share edit <share_id> [flags]
```

### Options

```
      --description string   Share description
      --editor               Use a text editor to define parameters
  -h, --help                 help for edit
      --name string          Share name
      --new-size int         New share size in GB
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
      --region string          Region (skip region discovery if set)
```

### SEE ALSO

* [ovhcloud cloud storage file share](ovhcloud_cloud_storage_file_share.md)	 - Manage file storage shares

