## ovhcloud cloud managed-rancher

Manage Rancher services in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for managed-rancher
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud managed-rancher create](ovhcloud_cloud_managed-rancher_create.md)	 - Create a new Rancher service
* [ovhcloud cloud managed-rancher delete](ovhcloud_cloud_managed-rancher_delete.md)	 - Delete a specific Rancher service
* [ovhcloud cloud managed-rancher edit](ovhcloud_cloud_managed-rancher_edit.md)	 - Edit the given Rancher service
* [ovhcloud cloud managed-rancher get](ovhcloud_cloud_managed-rancher_get.md)	 - Get a specific Rancher service
* [ovhcloud cloud managed-rancher list](ovhcloud_cloud_managed-rancher_list.md)	 - List Rancher services
* [ovhcloud cloud managed-rancher plan](ovhcloud_cloud_managed-rancher_plan.md)	 - List available Rancher plans in the given cloud project
* [ovhcloud cloud managed-rancher reset-admin-credentials](ovhcloud_cloud_managed-rancher_reset-admin-credentials.md)	 - Reset admin user credentials
* [ovhcloud cloud managed-rancher version](ovhcloud_cloud_managed-rancher_version.md)	 - List available Rancher versions in the given cloud project

