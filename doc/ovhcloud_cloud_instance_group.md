## ovhcloud cloud instance group

Manage instance groups

### Options

```
  -h, --help   help for group
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

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project
* [ovhcloud cloud instance group create](ovhcloud_cloud_instance_group_create.md)	 - Create an instance group
* [ovhcloud cloud instance group delete](ovhcloud_cloud_instance_group_delete.md)	 - Delete a specific instance group
* [ovhcloud cloud instance group get](ovhcloud_cloud_instance_group_get.md)	 - Get a specific instance group
* [ovhcloud cloud instance group list](ovhcloud_cloud_instance_group_list.md)	 - List all instance groups in the current cloud project

