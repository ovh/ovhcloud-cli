## ovhcloud cloud ip extnet

Manage ext-net public IPs in the given cloud project

### Options

```
  -h, --help   help for extnet
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
                               
                               When extracting a single scalar field, the value is printed without surrounding
                               quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud ip](ovhcloud_cloud_ip.md)	 - Manage public IPs (floating, additional and ext-net) in the given cloud project
* [ovhcloud cloud ip extnet delete](ovhcloud_cloud_ip_extnet_delete.md)	 - Delete a specific ext-net IP
* [ovhcloud cloud ip extnet get](ovhcloud_cloud_ip_extnet_get.md)	 - Get a specific ext-net IP
* [ovhcloud cloud ip extnet list](ovhcloud_cloud_ip_extnet_list.md)	 - List ext-net IPs

