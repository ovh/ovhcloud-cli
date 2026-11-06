## ovhcloud cloud instance snapshot create

Create a snapshot of the given instance

```
ovhcloud cloud instance snapshot create <instance_id> <snapshot_name> [flags]
```

### Options

```
      --distant-region-name string     Name of the distant region (for cross region snapshot)
      --distant-snapshot-name string   Name of the snapshot in the distant region (for cross region snapshot)
  -h, --help                           help for create
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

* [ovhcloud cloud instance snapshot](ovhcloud_cloud_instance_snapshot.md)	 - Manage snapshots of the given instance

