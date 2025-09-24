## ovhcloud cloud network private region

Manage regions in a specific private network

### Options

```
  -h, --help   help for region
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

* [ovhcloud cloud network private](ovhcloud_cloud_network_private.md)	 - Manage private networks in the given cloud project
* [ovhcloud cloud network private region add](ovhcloud_cloud_network_private_region_add.md)	 - Add a region to a private network
* [ovhcloud cloud network private region delete](ovhcloud_cloud_network_private_region_delete.md)	 - Delete the given region from a private network

