## ovhcloud cloud key-manager container

Manage Key Manager containers

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

* [ovhcloud cloud key-manager](ovhcloud_cloud_key-manager.md)	 - Manage Key Management Service (KMS) resources in the given cloud project
* [ovhcloud cloud key-manager container consumer](ovhcloud_cloud_key-manager_container_consumer.md)	 - Manage consumers of a Key Manager container
* [ovhcloud cloud key-manager container create](ovhcloud_cloud_key-manager_container_create.md)	 - Create a new Key Manager container
* [ovhcloud cloud key-manager container delete](ovhcloud_cloud_key-manager_container_delete.md)	 - Delete the given Key Manager container
* [ovhcloud cloud key-manager container edit](ovhcloud_cloud_key-manager_container_edit.md)	 - Edit the given Key Manager container (only secret references are mutable)
* [ovhcloud cloud key-manager container get](ovhcloud_cloud_key-manager_container_get.md)	 - Get a specific Key Manager container
* [ovhcloud cloud key-manager container list](ovhcloud_cloud_key-manager_container_list.md)	 - List Key Manager containers

