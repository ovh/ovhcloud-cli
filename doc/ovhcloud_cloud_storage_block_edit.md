## ovhcloud cloud storage block edit

Edit the given volume

```
ovhcloud cloud storage block edit <volume_id> [flags]
```

### Options

```
      --editor        Use a text editor to define parameters
  -h, --help          help for edit
      --name string   Volume name
      --size int      Volume size (in GB, can only be increased)
      --type string   Volume type (CLASSIC, HIGH_SPEED, HIGH_SPEED_GEN2)
      --wait          Wait for the volume to be READY before exiting
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

* [ovhcloud cloud storage block](ovhcloud_cloud_storage_block.md)	 - Manage block storage volumes in the given cloud project

