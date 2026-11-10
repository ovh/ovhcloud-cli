## ovhcloud cloud managed-kubernetes nodepool edit

Edit the given Kubernetes node pool

### Synopsis

Edit the given Kubernetes node pool. Only the node count, the autoscaler and the node template can be changed: the name, flavor, availability zone, billing type, anti-affinity and floating IP settings are immutable.

```
ovhcloud cloud managed-kubernetes nodepool edit <cluster_id> <nodepool_id> [flags]
```

### Options

```
      --autoscaler-enabled                                  Enable node autoscaling
      --autoscaler-max-nodes int                            Maximum number of nodes the autoscaler provisions
      --autoscaler-min-nodes int                            Minimum number of nodes the autoscaler maintains
      --autoscaler-scale-down-unneeded-time-seconds int     Time in seconds a node must stay below the utilization threshold before it is removed
      --autoscaler-scale-down-unready-time-seconds int      Time in seconds an unready node is kept before it is removed
      --autoscaler-scale-down-utilization-threshold float   Node utilization ratio (between 0 and 1) below which a node becomes a scale-down candidate
      --desired-nodes int                                   Desired number of nodes
      --editor                                              Use a text editor to define parameters
  -h, --help                                                help for edit
      --template-annotations stringToString                 Annotations to apply to each node (default [])
      --template-finalizers strings                         Finalizers to apply to each node
      --template-labels stringToString                      Labels to apply to each node (default [])
      --template-taints strings                             Taints to apply to each node in key[=value]:effect format (effect: NoExecute, NoSchedule, PreferNoSchedule)
      --template-unschedulable                              Mark the nodes as unschedulable
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

