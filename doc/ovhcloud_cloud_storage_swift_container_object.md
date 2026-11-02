## ovhcloud cloud storage swift container object

Manage objects of the given Swift container

### Options

```
  -h, --help   help for object
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

* [ovhcloud cloud storage swift container](ovhcloud_cloud_storage_swift_container.md)	 - Manage Swift storage containers in the given cloud project
* [ovhcloud cloud storage swift container object delete](ovhcloud_cloud_storage_swift_container_object_delete.md)	 - Delete the given object from the given Swift container
* [ovhcloud cloud storage swift container object get](ovhcloud_cloud_storage_swift_container_object_get.md)	 - Get a specific object of the given Swift container
* [ovhcloud cloud storage swift container object list](ovhcloud_cloud_storage_swift_container_object_list.md)	 - List objects of the given Swift container

