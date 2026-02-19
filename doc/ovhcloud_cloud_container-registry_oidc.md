## ovhcloud cloud container-registry oidc

Manage container registry OIDC integration

### Options

```
  -h, --help   help for oidc
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

* [ovhcloud cloud container-registry](ovhcloud_cloud_container-registry.md)	 - Manage container registries in the given cloud project
* [ovhcloud cloud container-registry oidc create](ovhcloud_cloud_container-registry_oidc_create.md)	 - Create a new OIDC configuration for a container registry
* [ovhcloud cloud container-registry oidc delete](ovhcloud_cloud_container-registry_oidc_delete.md)	 - Delete the OIDC configuration for a container registry
* [ovhcloud cloud container-registry oidc edit](ovhcloud_cloud_container-registry_oidc_edit.md)	 - Edit the OIDC configuration for a container registry
* [ovhcloud cloud container-registry oidc get](ovhcloud_cloud_container-registry_oidc_get.md)	 - Get OIDC configuration for a container registry

