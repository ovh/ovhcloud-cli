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
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud kube](ovhcloud_cloud_kube.md)	 - Manage Kubernetes clusters in the given cloud project
* [ovhcloud cloud kube oidc create](ovhcloud_cloud_kube_oidc_create.md)	 - Create a new OIDC integration for the given Kubernetes cluster
* [ovhcloud cloud kube oidc delete](ovhcloud_cloud_kube_oidc_delete.md)	 - Delete the OIDC integration for the given Kubernetes cluster
* [ovhcloud cloud kube oidc edit](ovhcloud_cloud_kube_oidc_edit.md)	 - Edit the OIDC configuration for the given Kubernetes cluster
* [ovhcloud cloud kube oidc get](ovhcloud_cloud_kube_oidc_get.md)	 - Get the OIDC configuration for the given Kubernetes cluster

