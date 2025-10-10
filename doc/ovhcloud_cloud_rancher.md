## ovhcloud cloud rancher

Manage Rancher services in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for rancher
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud rancher create](ovhcloud_cloud_rancher_create.md)	 - Create a new Rancher service
* [ovhcloud cloud rancher delete](ovhcloud_cloud_rancher_delete.md)	 - Delete a specific Rancher service
* [ovhcloud cloud rancher edit](ovhcloud_cloud_rancher_edit.md)	 - Edit the given Rancher service
* [ovhcloud cloud rancher get](ovhcloud_cloud_rancher_get.md)	 - Get a specific Rancher service
* [ovhcloud cloud rancher list](ovhcloud_cloud_rancher_list.md)	 - List Rancher services
* [ovhcloud cloud rancher reset-admin-credentials](ovhcloud_cloud_rancher_reset-admin-credentials.md)	 - Reset admin user credentials

