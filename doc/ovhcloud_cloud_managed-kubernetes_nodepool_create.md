## ovhcloud cloud managed-kubernetes nodepool create

Create a new Kubernetes node pool

### Synopsis

Use this command to create a node pool in the given managed Kubernetes cluster.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --flavor b3-16 --desired-nodes 3 --name newnodepool

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud managed-kubernetes nodepool create <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --from-file ./params.json --name NameOverriden

  It is also possible to use the interactive flavor selector to define the flavor parameter, like the following:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --init-file ./params.json --flavor-selector

3. Using your default text editor:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --editor --flavor b3-16

  You can also use the interactive flavor selector to define the flavor parameter, like the following:

	ovhcloud cloud managed-kubernetes nodepool create <cluster_id> --editor --flavor-selector


```
ovhcloud cloud managed-kubernetes nodepool create <cluster_id> [flags]
```

### Options

```
      --anti-affinity-enabled                               Spread the nodes across distinct hypervisors, capping the node pool size to 5
      --attach-floating-ip                                  Attach a floating IP to each node
      --autoscaler-enabled                                  Enable node autoscaling
      --autoscaler-max-nodes int                            Maximum number of nodes the autoscaler provisions
      --autoscaler-min-nodes int                            Minimum number of nodes the autoscaler maintains
      --autoscaler-scale-down-unneeded-time-seconds int     Time in seconds a node must stay below the utilization threshold before it is removed
      --autoscaler-scale-down-unready-time-seconds int      Time in seconds an unready node is kept before it is removed
      --autoscaler-scale-down-utilization-threshold float   Node utilization ratio (between 0 and 1) below which a node becomes a scale-down candidate
      --availability-zone string                            Availability zone the nodes are placed in, only available in 3-AZ regions
      --billing-type string                                 Billing period of the nodes (HOURLY, MONTHLY)
      --desired-nodes int                                   Desired number of nodes
      --editor                                              Use a text editor to define parameters
      --flavor string                                       Flavor of the nodes, e.g. b3-8, b3-16 (required)
      --flavor-selector                                     Use the interactive flavor selector
      --from-file string                                    File containing parameters
  -h, --help                                                help for create
      --init-file string                                    Create a file with example parameters
      --name string                                         Name of the node pool
      --replace                                             Replace parameters file if it already exists
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

