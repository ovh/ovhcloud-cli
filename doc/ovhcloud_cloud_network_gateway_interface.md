## ovhcloud cloud network gateway interface

Manage interfaces of a specific gateway

### Options

```
  -h, --help   help for interface
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

* [ovhcloud cloud network gateway](ovhcloud_cloud_network_gateway.md)	 - Manage gateways in the given cloud project
* [ovhcloud cloud network gateway interface create](ovhcloud_cloud_network_gateway_interface_create.md)	 - Create a new interface for the given gateway
* [ovhcloud cloud network gateway interface delete](ovhcloud_cloud_network_gateway_interface_delete.md)	 - Delete a specific interface of a gateway
* [ovhcloud cloud network gateway interface get](ovhcloud_cloud_network_gateway_interface_get.md)	 - Get a specific interface of a gateway
* [ovhcloud cloud network gateway interface list](ovhcloud_cloud_network_gateway_interface_list.md)	 - List interfaces of a specific gateway

