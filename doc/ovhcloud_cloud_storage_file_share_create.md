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
      --protocol string            Share protocol (default "NFS")
      --replace                    Replace parameters file if it already exists
      --share-network-id string    Share network ID
      --share-type string          Share type (default "STANDARD_1AZ")
      --size int                   Share size in GB
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

* [ovhcloud cloud storage file share](ovhcloud_cloud_storage_file_share.md)	 - Manage file storage shares

