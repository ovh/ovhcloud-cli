## ovhcloud cloud kube nodepool edit

Edit the given Kubernetes node pool

```
ovhcloud cloud kube nodepool edit <cluster_id> <nodepool_id> [flags]
```

### Options

```
      --autoscale                                Enable autoscaling for the node pool
      --desired-nodes int                        Desired number of nodes
      --editor                                   Use a text editor to define parameters
  -h, --help                                     help for edit
      --max-nodes int                            Higher limit you accept for the desiredNodes value (100 by default)
      --min-nodes int                            Lower limit you accept for the desiredNodes value (0 by default)
      --nodes-to-remove strings                  List of node IDs to remove from the node pool
      --scale-down-unneeded-time-seconds int     How long a node should be unneeded before it is eligible for scale down (seconds)
      --scale-down-unready-time-seconds int      How long an unready node should be unneeded before it is eligible for scale down (seconds)
      --scale-down-utilization-threshold float   Sum of CPU or memory of all pods running on the node divided by node's corresponding allocatable resource, below which a node can be considered for scale down
      --template-annotations stringToString      Annotations to apply to each node (default [])
      --template-finalizers strings              Finalizers to apply to each node
      --template-labels stringToString           Labels to apply to each node (default [])
      --template-taints strings                  Taints to apply to each node in key=value:effect format
      --template-unschedulable                   Set the nodes as unschedulable
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud kube nodepool](ovhcloud_cloud_kube_nodepool.md)	 - Manage Kubernetes node pools

