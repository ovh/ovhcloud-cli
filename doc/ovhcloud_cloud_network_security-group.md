## ovhcloud cloud network security-group

Manage security groups in the given cloud project

### Options

```
  -h, --help   help for security-group
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

* [ovhcloud cloud network](ovhcloud_cloud_network.md)	 - Manage networks in the given cloud project
* [ovhcloud cloud network security-group create](ovhcloud_cloud_network_security-group_create.md)	 - Create a new security group
* [ovhcloud cloud network security-group delete](ovhcloud_cloud_network_security-group_delete.md)	 - Delete a specific security group
* [ovhcloud cloud network security-group edit](ovhcloud_cloud_network_security-group_edit.md)	 - Edit the given security group
* [ovhcloud cloud network security-group get](ovhcloud_cloud_network_security-group_get.md)	 - Get a specific security group
* [ovhcloud cloud network security-group list](ovhcloud_cloud_network_security-group_list.md)	 - List security groups

