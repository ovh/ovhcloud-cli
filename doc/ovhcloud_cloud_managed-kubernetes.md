## ovhcloud cloud managed-kubernetes

Manage Kubernetes clusters in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for managed-kubernetes
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud managed-kubernetes create](ovhcloud_cloud_managed-kubernetes_create.md)	 - Create a new Kubernetes cluster
* [ovhcloud cloud managed-kubernetes delete](ovhcloud_cloud_managed-kubernetes_delete.md)	 - Delete the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes edit](ovhcloud_cloud_managed-kubernetes_edit.md)	 - Edit the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes event](ovhcloud_cloud_managed-kubernetes_event.md)	 - Inspect the events of a Kubernetes cluster
* [ovhcloud cloud managed-kubernetes flavor](ovhcloud_cloud_managed-kubernetes_flavor.md)	 - Explore the Kubernetes node flavors
* [ovhcloud cloud managed-kubernetes get](ovhcloud_cloud_managed-kubernetes_get.md)	 - Get the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes kubeconfig](ovhcloud_cloud_managed-kubernetes_kubeconfig.md)	 - Manage the kubeconfig for the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes list](ovhcloud_cloud_managed-kubernetes_list.md)	 - List your Kubernetes clusters
* [ovhcloud cloud managed-kubernetes node](ovhcloud_cloud_managed-kubernetes_node.md)	 - Inspect Kubernetes nodes
* [ovhcloud cloud managed-kubernetes nodepool](ovhcloud_cloud_managed-kubernetes_nodepool.md)	 - Manage Kubernetes node pools
* [ovhcloud cloud managed-kubernetes plan](ovhcloud_cloud_managed-kubernetes_plan.md)	 - Explore the Kubernetes cluster plans
* [ovhcloud cloud managed-kubernetes region](ovhcloud_cloud_managed-kubernetes_region.md)	 - Explore the Kubernetes regions
* [ovhcloud cloud managed-kubernetes restart](ovhcloud_cloud_managed-kubernetes_restart.md)	 - Restart control plane apiserver to invalidate cache without downtime
* [ovhcloud cloud managed-kubernetes task](ovhcloud_cloud_managed-kubernetes_task.md)	 - Inspect the asynchronous tasks of a Kubernetes cluster
* [ovhcloud cloud managed-kubernetes usage](ovhcloud_cloud_managed-kubernetes_usage.md)	 - Inspect the resource usage of a Kubernetes cluster
* [ovhcloud cloud managed-kubernetes version](ovhcloud_cloud_managed-kubernetes_version.md)	 - Explore the Kubernetes versions

