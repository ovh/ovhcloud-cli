## ovhcloud cloud loadbalancer associate-floating-ip

Associate an existing floating IP to a loadbalancer

```
ovhcloud cloud loadbalancer associate-floating-ip <loadbalancer_id> [flags]
```

### Options

```
      --editor                  Use a text editor to define parameters
      --floating-ip-id string   Floating IP ID
      --from-file string        File containing parameters
  -h, --help                    help for associate-floating-ip
      --init-file string        Create a file with example parameters
      --ip string               Private loadbalancer IP to associate the floating IP with
      --replace                 Replace parameters file if it already exists
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

* [ovhcloud cloud loadbalancer](ovhcloud_cloud_loadbalancer.md)	 - Manage loadbalancers in the given cloud project

