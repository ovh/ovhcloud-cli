## ovhcloud iploadbalancing edit

Edit the given IpLoadbalancing

```
ovhcloud iploadbalancing edit <service_name> [flags]
```

### Options

```
      --display-name string        Display name of the load balancer
      --editor                     Use a text editor to define parameters
  -h, --help                       help for edit
      --ssl-configuration string   SSL configuration of the load balancer (intermediate, modern)
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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

* [ovhcloud iploadbalancing](ovhcloud_iploadbalancing.md)	 - Retrieve information and manage your IP LoadBalancing services

