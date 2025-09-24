## ovhcloud cloud kube nodepool create

Create a new Kubernetes node pool

### Synopsis

Use this command to create a node pool in the given managed Kubernetes cluster.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud kube nodepool create <cluster_id> --flavor-name b3-16 --desired-nodes 3 --name newnodepool

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud kube nodepool create <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud kube nodepool create <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud kube nodepool create <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud kube nodepool create <cluster_id> --from-file ./params.json --name NameOverriden

  It is also possible to use the interactive flavor selector to define the flavor-name parameter, like the following:

	ovhcloud cloud kube nodepool create <cluster_id> --init-file ./params.json --flavor-selector

3. Using your default text editor:

	ovhcloud cloud kube nodepool create <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud kube nodepool create <cluster_id> --editor --flavor-name b3-16

  You can also use the interactive flavor selector to define the flavor-name parameter, like the following:

	ovhcloud cloud kube nodepool create <cluster_id> --editor --flavor-selector


```
ovhcloud cloud kube nodepool create <cluster_id> [flags]
```

### Options

```
      --anti-affinity                            Enable anti-affinity for the node pool
      --autoscale                                Enable autoscaling for the node pool
      --availability-zones strings               Availability zones for the node pool
      --desired-nodes int                        Desired number of nodes
      --editor                                   Use a text editor to define parameters
      --flavor-name string                       Flavor name for the nodes (b2-7, b2-15, etc.)
      --flavor-selector                          Use the interactive flavor selector
      --from-file string                         File containing parameters
  -h, --help                                     help for create
      --init-file string                         Create a file with example parameters
      --max-nodes int                            Higher limit you accept for the desiredNodes value (100 by default)
      --min-nodes int                            Lower limit you accept for the desiredNodes value (0 by default)
      --monthly-billed                           Enable monthly billing for the node pool
      --name string                              Name of the node pool
      --replace                                  Replace parameters file if it already exists
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

