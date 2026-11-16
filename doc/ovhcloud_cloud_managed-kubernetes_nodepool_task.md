## ovhcloud cloud managed-kubernetes nodepool task

Inspect the asynchronous tasks of a Kubernetes node pool

### Options

```
  -h, --help   help for task
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

* [ovhcloud cloud managed-kubernetes nodepool](ovhcloud_cloud_managed-kubernetes_nodepool.md)	 - Manage Kubernetes node pools
* [ovhcloud cloud managed-kubernetes nodepool task get](ovhcloud_cloud_managed-kubernetes_nodepool_task_get.md)	 - Get the given Kubernetes node pool task
* [ovhcloud cloud managed-kubernetes nodepool task list](ovhcloud_cloud_managed-kubernetes_nodepool_task_list.md)	 - List the tasks of the given Kubernetes node pool

