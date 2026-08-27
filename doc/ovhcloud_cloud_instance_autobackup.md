## ovhcloud cloud instance autobackup

Manage automatic backup workflows for instances

### Options

```
  -h, --help   help for autobackup
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

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project
* [ovhcloud cloud instance autobackup create](ovhcloud_cloud_instance_autobackup_create.md)	 - Create an automatic backup workflow for the given instance
* [ovhcloud cloud instance autobackup delete](ovhcloud_cloud_instance_autobackup_delete.md)	 - Delete an automatic backup workflow
* [ovhcloud cloud instance autobackup get](ovhcloud_cloud_instance_autobackup_get.md)	 - Get details of an automatic backup workflow
* [ovhcloud cloud instance autobackup list](ovhcloud_cloud_instance_autobackup_list.md)	 - List automatic backup workflows for the given instance

