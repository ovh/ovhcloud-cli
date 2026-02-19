## ovhcloud cloud network private subnet

Manage subnets in a specific private network

### Options

```
  -h, --help   help for subnet
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

* [ovhcloud cloud network private](ovhcloud_cloud_network_private.md)	 - Manage private networks in the given cloud project
* [ovhcloud cloud network private subnet create](ovhcloud_cloud_network_private_subnet_create.md)	 - Create a subnet in the given private network
* [ovhcloud cloud network private subnet delete](ovhcloud_cloud_network_private_subnet_delete.md)	 - Delete a specific subnet in a private network
* [ovhcloud cloud network private subnet edit](ovhcloud_cloud_network_private_subnet_edit.md)	 - Edit a specific subnet in a private network
* [ovhcloud cloud network private subnet get](ovhcloud_cloud_network_private_subnet_get.md)	 - Get a specific subnet in a private network
* [ovhcloud cloud network private subnet list](ovhcloud_cloud_network_private_subnet_list.md)	 - List subnets in a private network

