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
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --output json
                                 --output yaml
                                 --output interactive
                                 --output 'id' (to extract a single field)
                                 --output 'nested.field.subfield' (to extract a nested field)
                                 --output '[id, "name"]' (to extract multiple fields as an array)
                                 --output '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --output 'name+","+type' (to extract and concatenate fields in a string)
                                 --output '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
                               
                               When extracting a single scalar field, the value is printed without surrounding
                               quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string         Use a specific profile from the configuration file
      --region string          Region (skip region discovery if set)
```

### SEE ALSO

* [ovhcloud cloud storage file share](ovhcloud_cloud_storage_file_share.md)	 - Manage file storage shares

