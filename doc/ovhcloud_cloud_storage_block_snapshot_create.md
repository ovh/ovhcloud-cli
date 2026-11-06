## ovhcloud cloud storage block snapshot create

Create a snapshot of the given volume

```
ovhcloud cloud storage block snapshot create <volume_id> [flags]
```

### Options

```
      --description string   Snapshot description
  -h, --help                 help for create
      --name string          Snapshot name
      --wait                 Wait for the snapshot to be READY before exiting
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

* [ovhcloud cloud storage block snapshot](ovhcloud_cloud_storage_block_snapshot.md)	 - Manage snapshots of the given volume

