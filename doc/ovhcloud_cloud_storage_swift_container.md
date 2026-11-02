## ovhcloud cloud storage swift container

Manage Swift storage containers in the given cloud project

### Options

```
  -h, --help   help for container
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

* [ovhcloud cloud storage swift](ovhcloud_cloud_storage_swift.md)	 - Manage Swift storage in the given cloud project
* [ovhcloud cloud storage swift container create](ovhcloud_cloud_storage_swift_container_create.md)	 - Create a new Swift container
* [ovhcloud cloud storage swift container delete](ovhcloud_cloud_storage_swift_container_delete.md)	 - Delete the given Swift container and all its objects
* [ovhcloud cloud storage swift container edit](ovhcloud_cloud_storage_swift_container_edit.md)	 - Edit the given Swift container
* [ovhcloud cloud storage swift container get](ovhcloud_cloud_storage_swift_container_get.md)	 - Get a specific Swift container
* [ovhcloud cloud storage swift container list](ovhcloud_cloud_storage_swift_container_list.md)	 - List Swift containers
* [ovhcloud cloud storage swift container object](ovhcloud_cloud_storage_swift_container_object.md)	 - Manage objects of the given Swift container
* [ovhcloud cloud storage swift container presign](ovhcloud_cloud_storage_swift_container_presign.md)	 - Generate a presigned URL to download an object of the given Swift container

