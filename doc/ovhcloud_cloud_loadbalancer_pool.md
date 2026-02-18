## ovhcloud cloud loadbalancer pool

Manage pools of loadbalancers

### Options

```
  -h, --help   help for pool
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
* [ovhcloud cloud loadbalancer pool create](ovhcloud_cloud_loadbalancer_pool_create.md)	 - Create a pool in the given region
* [ovhcloud cloud loadbalancer pool delete](ovhcloud_cloud_loadbalancer_pool_delete.md)	 - Delete a specific pool
* [ovhcloud cloud loadbalancer pool edit](ovhcloud_cloud_loadbalancer_pool_edit.md)	 - Edit a specific pool
* [ovhcloud cloud loadbalancer pool get](ovhcloud_cloud_loadbalancer_pool_get.md)	 - Get a specific pool
* [ovhcloud cloud loadbalancer pool list](ovhcloud_cloud_loadbalancer_pool_list.md)	 - List all pools
* [ovhcloud cloud loadbalancer pool member](ovhcloud_cloud_loadbalancer_pool_member.md)	 - Manage members of a loadbalancer pool

