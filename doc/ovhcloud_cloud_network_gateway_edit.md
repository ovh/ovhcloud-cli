## ovhcloud cloud network gateway edit

Edit the given gateway

```
ovhcloud cloud network gateway edit <gateway_id> [flags]
```

### Options

```
      --description string              Description of the gateway
      --editor                          Use a text editor to define parameters
      --external-gateway-enabled        Whether the external gateway is enabled
      --external-gateway-model string   External gateway sizing model (S, M, L, XL, 2XL, 3XL)
  -h, --help                            help for edit
      --name string                     Name of the gateway
      --subnet strings                  ID of a subnet to attach to the gateway (repeatable)
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

* [ovhcloud cloud network gateway](ovhcloud_cloud_network_gateway.md)	 - Manage gateways in the given cloud project

