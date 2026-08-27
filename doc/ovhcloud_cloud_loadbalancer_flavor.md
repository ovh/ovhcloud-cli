## ovhcloud cloud loadbalancer flavor

List available loadbalancer flavors in the given cloud project

### Options

```
  -h, --help   help for flavor
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

* [ovhcloud cloud loadbalancer](ovhcloud_cloud_loadbalancer.md)	 - Manage loadbalancers in the given cloud project
* [ovhcloud cloud loadbalancer flavor get](ovhcloud_cloud_loadbalancer_flavor_get.md)	 - Get details of a specific loadbalancer flavor
* [ovhcloud cloud loadbalancer flavor list](ovhcloud_cloud_loadbalancer_flavor_list.md)	 - List available loadbalancer flavors in the given cloud project

