## ovhcloud cloud storage file snapshot create

Create a snapshot of the given share

```
ovhcloud cloud storage file snapshot create <share_id> [flags]
```

### Options

```
      --description string   Snapshot description
      --editor               Use a text editor to define parameters
      --from-file string     File containing parameters
  -h, --help                 help for create
      --init-file string     Create a file with example parameters
      --name string          Snapshot name
      --replace              Replace parameters file if it already exists
      --wait                 Wait for the snapshot to be ready before exiting
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

* [ovhcloud cloud storage file snapshot](ovhcloud_cloud_storage_file_snapshot.md)	 - Manage share snapshots

