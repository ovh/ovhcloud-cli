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
* [ovhcloud cloud managed-kubernetes customization](ovhcloud_cloud_managed-kubernetes_customization.md)	 - Manage Kubernetes cluster customizations
* [ovhcloud cloud managed-kubernetes delete](ovhcloud_cloud_managed-kubernetes_delete.md)	 - Delete the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes edit](ovhcloud_cloud_managed-kubernetes_edit.md)	 - Edit the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes get](ovhcloud_cloud_managed-kubernetes_get.md)	 - Get the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes ip-restrictions](ovhcloud_cloud_managed-kubernetes_ip-restrictions.md)	 - Manage IP restrictions for Kubernetes clusters
* [ovhcloud cloud managed-kubernetes kubeconfig](ovhcloud_cloud_managed-kubernetes_kubeconfig.md)	 - Manage the kubeconfig for the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes list](ovhcloud_cloud_managed-kubernetes_list.md)	 - List your Kubernetes clusters
* [ovhcloud cloud managed-kubernetes node](ovhcloud_cloud_managed-kubernetes_node.md)	 - Manage Kubernetes nodes
* [ovhcloud cloud managed-kubernetes nodepool](ovhcloud_cloud_managed-kubernetes_nodepool.md)	 - Manage Kubernetes node pools
* [ovhcloud cloud managed-kubernetes oidc](ovhcloud_cloud_managed-kubernetes_oidc.md)	 - Manage OpenID Connect (OIDC) integration for Kubernetes clusters
* [ovhcloud cloud managed-kubernetes private-network-configuration](ovhcloud_cloud_managed-kubernetes_private-network-configuration.md)	 - Manage private network configuration for Kubernetes clusters
* [ovhcloud cloud managed-kubernetes reset](ovhcloud_cloud_managed-kubernetes_reset.md)	 - Reset the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes restart](ovhcloud_cloud_managed-kubernetes_restart.md)	 - Restart control plane apiserver to invalidate cache without downtime
* [ovhcloud cloud managed-kubernetes set-load-balancers-subnet](ovhcloud_cloud_managed-kubernetes_set-load-balancers-subnet.md)	 - Update the load balancers subnet ID for the given Kubernetes cluster
* [ovhcloud cloud managed-kubernetes update](ovhcloud_cloud_managed-kubernetes_update.md)	 - Update the given Kubernetes cluster

