## ovhcloud cloud network private

Manage private networks in the given cloud project

### Options

```
  -h, --help   help for private
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

* [ovhcloud cloud network](ovhcloud_cloud_network.md)	 - Manage networks in the given cloud project
* [ovhcloud cloud network private create](ovhcloud_cloud_network_private_create.md)	 - Create a private network in the given cloud project
* [ovhcloud cloud network private delete](ovhcloud_cloud_network_private_delete.md)	 - Delete a specific private network
* [ovhcloud cloud network private edit](ovhcloud_cloud_network_private_edit.md)	 - Edit the given private network
* [ovhcloud cloud network private get](ovhcloud_cloud_network_private_get.md)	 - Get a specific private network
* [ovhcloud cloud network private list](ovhcloud_cloud_network_private_list.md)	 - List your private networks
* [ovhcloud cloud network private region](ovhcloud_cloud_network_private_region.md)	 - Manage regions in a specific private network
* [ovhcloud cloud network private subnet](ovhcloud_cloud_network_private_subnet.md)	 - Manage subnets in a specific private network

