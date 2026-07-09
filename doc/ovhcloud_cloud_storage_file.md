## ovhcloud cloud storage file

Manage file storage in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for file
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud storage](ovhcloud_cloud_storage.md)	 - Manage storage services in the given cloud project
* [ovhcloud cloud storage file network](ovhcloud_cloud_storage_file_network.md)	 - Manage file storage share networks
* [ovhcloud cloud storage file share](ovhcloud_cloud_storage_file_share.md)	 - Manage file storage shares
* [ovhcloud cloud storage file snapshot](ovhcloud_cloud_storage_file_snapshot.md)	 - Manage file storage snapshots

