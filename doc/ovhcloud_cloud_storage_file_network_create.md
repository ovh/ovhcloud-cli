## ovhcloud cloud storage file network create

Create a new file storage share network

```
ovhcloud cloud storage file network create [flags]
```

### Options

```
      --availability-zone string   Availability zone within the region
      --description string         Description of the share network
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Name of the share network
      --network-id string          ID of the private network to back the share network
      --region string              Region where the share network is created
      --replace                    Replace parameters file if it already exists
      --subnet-id string           ID of the subnet to back the share network
      --wait                       Wait for the share network to be ready before exiting
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

* [ovhcloud cloud storage file network](ovhcloud_cloud_storage_file_network.md)	 - Manage file storage share networks

