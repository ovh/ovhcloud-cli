## ovhcloud cloud container-registry ip-restrictions registry

Manage IP restrictions for container registry artifact manager (Docker, Helm...) access

### Options

```
  -h, --help   help for registry
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
```

### SEE ALSO

* [ovhcloud cloud container-registry ip-restrictions](ovhcloud_cloud_container-registry_ip-restrictions.md)	 - Manage container registry IP restrictions
* [ovhcloud cloud container-registry ip-restrictions registry add](ovhcloud_cloud_container-registry_ip-restrictions_registry_add.md)	 - Add a registry IP restriction to a container registry
* [ovhcloud cloud container-registry ip-restrictions registry delete](ovhcloud_cloud_container-registry_ip-restrictions_registry_delete.md)	 - Delete a registry IP restriction from a container registry
* [ovhcloud cloud container-registry ip-restrictions registry list](ovhcloud_cloud_container-registry_ip-restrictions_registry_list.md)	 - List registry IP restrictions for a container registry

