## ovhcloud storage-netapp

Retrieve information and manage your Storage NetApp services

### Options

```
  -h, --help   help for storage-netapp
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud storage-netapp edit](ovhcloud_storage-netapp_edit.md)	 - Edit the given StorageNetApp
* [ovhcloud storage-netapp get](ovhcloud_storage-netapp_get.md)	 - Retrieve information of a specific StorageNetApp
* [ovhcloud storage-netapp list](ovhcloud_storage-netapp_list.md)	 - List your Storage NetApp services

