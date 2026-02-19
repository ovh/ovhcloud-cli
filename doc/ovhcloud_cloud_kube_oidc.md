## ovhcloud cloud kube oidc

Manage OpenID Connect (OIDC) integration for Kubernetes clusters

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

* [ovhcloud cloud kube](ovhcloud_cloud_kube.md)	 - Manage Kubernetes clusters in the given cloud project
* [ovhcloud cloud kube oidc create](ovhcloud_cloud_kube_oidc_create.md)	 - Create a new OIDC integration for the given Kubernetes cluster
* [ovhcloud cloud kube oidc delete](ovhcloud_cloud_kube_oidc_delete.md)	 - Delete the OIDC integration for the given Kubernetes cluster
* [ovhcloud cloud kube oidc edit](ovhcloud_cloud_kube_oidc_edit.md)	 - Edit the OIDC configuration for the given Kubernetes cluster
* [ovhcloud cloud kube oidc get](ovhcloud_cloud_kube_oidc_get.md)	 - Get the OIDC configuration for the given Kubernetes cluster

