## ovhcloud cloud loadbalancer pool member

Manage members of a loadbalancer pool

### Options

```
  -h, --help   help for member
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

* [ovhcloud cloud loadbalancer pool](ovhcloud_cloud_loadbalancer_pool.md)	 - Manage pools of loadbalancers
* [ovhcloud cloud loadbalancer pool member create](ovhcloud_cloud_loadbalancer_pool_member_create.md)	 - Create member(s) in a specific pool
* [ovhcloud cloud loadbalancer pool member delete](ovhcloud_cloud_loadbalancer_pool_member_delete.md)	 - Delete a specific pool member
* [ovhcloud cloud loadbalancer pool member edit](ovhcloud_cloud_loadbalancer_pool_member_edit.md)	 - Edit a specific pool member
* [ovhcloud cloud loadbalancer pool member get](ovhcloud_cloud_loadbalancer_pool_member_get.md)	 - Get a specific pool member
* [ovhcloud cloud loadbalancer pool member list](ovhcloud_cloud_loadbalancer_pool_member_list.md)	 - List members of a specific pool

