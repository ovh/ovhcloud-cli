## ovhcloud cloud container-registry ip-restrictions management

Manage IP restrictions for container registry Harbor UI and API access

### Options

```
  -h, --help   help for management
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud container-registry ip-restrictions](ovhcloud_cloud_container-registry_ip-restrictions.md)	 - Manage container registry IP restrictions
* [ovhcloud cloud container-registry ip-restrictions management add](ovhcloud_cloud_container-registry_ip-restrictions_management_add.md)	 - Add a management IP restriction to a container registry
* [ovhcloud cloud container-registry ip-restrictions management delete](ovhcloud_cloud_container-registry_ip-restrictions_management_delete.md)	 - Delete a management IP restriction from a container registry
* [ovhcloud cloud container-registry ip-restrictions management list](ovhcloud_cloud_container-registry_ip-restrictions_management_list.md)	 - List management IP restrictions for a container registry

