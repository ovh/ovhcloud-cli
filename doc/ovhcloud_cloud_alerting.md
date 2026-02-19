## ovhcloud cloud alerting

Manage billing alert configurations in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for alerting
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud alerting alert](ovhcloud_cloud_alerting_alert.md)	 - Manage triggered alerts for a billing alert configuration
* [ovhcloud cloud alerting create](ovhcloud_cloud_alerting_create.md)	 - Create a new billing alert configuration
* [ovhcloud cloud alerting delete](ovhcloud_cloud_alerting_delete.md)	 - Delete a billing alert configuration
* [ovhcloud cloud alerting edit](ovhcloud_cloud_alerting_edit.md)	 - Edit a billing alert configuration
* [ovhcloud cloud alerting get](ovhcloud_cloud_alerting_get.md)	 - Get a specific billing alert configuration
* [ovhcloud cloud alerting list](ovhcloud_cloud_alerting_list.md)	 - List billing alert configurations

