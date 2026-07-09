## ovhcloud cloud storage file network

Manage file storage share networks

### Options

```
  -h, --help   help for network
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

* [ovhcloud cloud storage file](ovhcloud_cloud_storage_file.md)	 - Manage file storage in the given cloud project
* [ovhcloud cloud storage file network create](ovhcloud_cloud_storage_file_network_create.md)	 - Create a new file storage share network
* [ovhcloud cloud storage file network delete](ovhcloud_cloud_storage_file_network_delete.md)	 - Delete the given file storage share network
* [ovhcloud cloud storage file network get](ovhcloud_cloud_storage_file_network_get.md)	 - Get a specific file storage share network
* [ovhcloud cloud storage file network list](ovhcloud_cloud_storage_file_network_list.md)	 - List file storage share networks

