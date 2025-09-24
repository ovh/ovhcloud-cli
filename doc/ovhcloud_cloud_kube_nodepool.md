## ovhcloud cloud kube nodepool

Manage Kubernetes node pools

### Options

```
  -h, --help   help for nodepool
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
* [ovhcloud cloud kube nodepool create](ovhcloud_cloud_kube_nodepool_create.md)	 - Create a new Kubernetes node pool
* [ovhcloud cloud kube nodepool delete](ovhcloud_cloud_kube_nodepool_delete.md)	 - Delete the given Kubernetes node pool
* [ovhcloud cloud kube nodepool edit](ovhcloud_cloud_kube_nodepool_edit.md)	 - Edit the given Kubernetes node pool
* [ovhcloud cloud kube nodepool get](ovhcloud_cloud_kube_nodepool_get.md)	 - Get the given Kubernetes node pool
* [ovhcloud cloud kube nodepool list](ovhcloud_cloud_kube_nodepool_list.md)	 - List node pools in the given Kubernetes cluster

