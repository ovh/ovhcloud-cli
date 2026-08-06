## ovhcloud cloud loadbalancer listener

Manage listeners of loadbalancers

### Options

```
  -h, --help   help for listener
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
  -y, --yes                    Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud loadbalancer](ovhcloud_cloud_loadbalancer.md)	 - Manage loadbalancers in the given cloud project
* [ovhcloud cloud loadbalancer listener create](ovhcloud_cloud_loadbalancer_listener_create.md)	 - Create a listener in the given region
* [ovhcloud cloud loadbalancer listener delete](ovhcloud_cloud_loadbalancer_listener_delete.md)	 - Delete a specific listener
* [ovhcloud cloud loadbalancer listener edit](ovhcloud_cloud_loadbalancer_listener_edit.md)	 - Edit a specific listener
* [ovhcloud cloud loadbalancer listener get](ovhcloud_cloud_loadbalancer_listener_get.md)	 - Get a specific listener
* [ovhcloud cloud loadbalancer listener list](ovhcloud_cloud_loadbalancer_listener_list.md)	 - List all listeners

