## ovhcloud cloud storage file share create

Create a new file storage share

```
ovhcloud cloud storage file share create [flags]
```

### Options

```
      --availability-zone string   Availability zone within the region
      --description string         Description of the file storage share
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Name of the file storage share
      --protocol string            File sharing protocol (NFS)
      --region string              Region where the share is created
      --replace                    Replace parameters file if it already exists
      --share-network-id string    ID of the share network to attach the share to
      --share-type string          File storage type / performance tier (STANDARD_1AZ)
      --size int                   Size of the file storage share in GB
      --wait                       Wait for the share to be ready before exiting
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud storage file share](ovhcloud_cloud_storage_file_share.md)	 - Manage file storage shares

