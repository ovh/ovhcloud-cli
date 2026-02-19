## ovhcloud cloud container-registry

Manage container registries in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for container-registry
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
* [ovhcloud cloud container-registry create](ovhcloud_cloud_container-registry_create.md)	 - Create a new container registry
* [ovhcloud cloud container-registry delete](ovhcloud_cloud_container-registry_delete.md)	 - Delete a specific container registry
* [ovhcloud cloud container-registry edit](ovhcloud_cloud_container-registry_edit.md)	 - Edit the given container registry
* [ovhcloud cloud container-registry get](ovhcloud_cloud_container-registry_get.md)	 - Get a specific container registry
* [ovhcloud cloud container-registry iam](ovhcloud_cloud_container-registry_iam.md)	 - Manage container registry IAM
* [ovhcloud cloud container-registry ip-restrictions](ovhcloud_cloud_container-registry_ip-restrictions.md)	 - Manage container registry IP restrictions
* [ovhcloud cloud container-registry list](ovhcloud_cloud_container-registry_list.md)	 - List your container registries
* [ovhcloud cloud container-registry oidc](ovhcloud_cloud_container-registry_oidc.md)	 - Manage container registry OIDC integration
* [ovhcloud cloud container-registry plan](ovhcloud_cloud_container-registry_plan.md)	 - Manage container registry plans
* [ovhcloud cloud container-registry users](ovhcloud_cloud_container-registry_users.md)	 - Manage container registry users

