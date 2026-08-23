## ovhcloud cloud storage file share create

Create a new share

```
ovhcloud cloud storage file share create <region> [flags]
```

### Options

```
      --availability-zone string   Availability zone (required in 3AZ regions)
      --description string         Share description
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Share name
      --network-id string          Network ID
      --replace                    Replace parameters file if it already exists
      --size int                   Share size in GB
      --snapshot-id string         Snapshot ID to create the share from
      --subnet-id string           Subnet ID
      --type string                Share type
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

