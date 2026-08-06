## ovhcloud cloud managed-kubernetes create

Create a new Kubernetes cluster

### Synopsis

Use this command to create a managed Kubernetes cluster in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-kubernetes create --name MyNewCluster --region SBG5 --version 1.32

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud managed-kubernetes create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud managed-kubernetes create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud managed-kubernetes create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud managed-kubernetes create --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud managed-kubernetes create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-kubernetes create --editor --region BHS5


```
ovhcloud cloud managed-kubernetes create [flags]
```

### Options

```
      --api-server.admission-plugins.disabled strings       Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)
      --api-server.admission-plugins.enabled strings        Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)
      --cilium-cluster-id uint8                             Cilium cluster ID
      --cilium-cluster-mesh-apiserver-node-port uint16      ClusterMesh API server node port
      --cilium-cluster-mesh-apiserver-service-type string   ClusterMesh API server service type (LoadBalancer, NodePort)
      --cilium-cluster-mesh-enabled                         Enable Cilium ClusterMesh
      --cilium-hubble-enabled                               Enable Hubble observability
      --cilium-hubble-relay-enabled                         Enable Hubble Relay
      --cilium-hubble-ui-backend-limits-cpu string          Hubble UI backend CPU limit (e.g. '500m')
      --cilium-hubble-ui-backend-limits-memory string       Hubble UI backend memory limit (e.g. '256Mi')
      --cilium-hubble-ui-backend-requests-cpu string        Hubble UI backend CPU request (e.g. '100m')
      --cilium-hubble-ui-backend-requests-memory string     Hubble UI backend memory request (e.g. '128Mi')
      --cilium-hubble-ui-enabled                            Enable Hubble UI
      --cilium-hubble-ui-frontend-limits-cpu string         Hubble UI frontend CPU limit (e.g. '500m')
      --cilium-hubble-ui-frontend-limits-memory string      Hubble UI frontend memory limit (e.g. '256Mi')
      --cilium-hubble-ui-frontend-requests-cpu string       Hubble UI frontend CPU request (e.g. '100m')
      --cilium-hubble-ui-frontend-requests-memory string    Hubble UI frontend memory request (e.g. '128Mi')
      --editor                                              Use a text editor to define parameters
      --from-file string                                    File containing parameters
  -h, --help                                                help for create
      --init-file string                                    Create a file with example parameters
      --ip-allocation-policy-pods-ipv4-cidr string          IPv4 CIDR for pods
      --ip-allocation-policy-services-ipv4-cidr string      IPv4 CIDR for services
      --kube-proxy-mode string                              Kube-proxy mode (iptables or ipvs)
      --kube-proxy.iptables.min-sync-period string          Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.iptables.sync-period string              Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.min-sync-period string              Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.scheduler string                    Scheduler for kube-proxy ipvs (dh, lc, nq, rr, sed, sh)
      --kube-proxy.ipvs.sync-period string                  Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.tcp-fin-timeout string              Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.tcp-timeout string                  Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.udp-timeout string                  Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')
      --load-balancers-subnet-id string                     OpenStack subnet ID that the load balancers will use
      --name string                                         Name of the Kubernetes cluster
      --nodes-subnet-id string                              OpenStack subnet ID that the cluster nodes will use
      --plan string                                         Kubernetes cluster plan (free or standard, default: free)
      --private-network-id string                           OpenStack private network ID that the cluster will use
      --private-network.default-vrack-gateway string        If defined, all egress traffic will be routed towards this IP address, which should belong to the private network
      --private-network.routing-as-default                  Set private network routing as default
      --region string                                       Region for the Kubernetes cluster
      --replace                                             Replace parameters file if it already exists
      --update-policy string                                Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)
      --version string                                      Kubernetes version
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
  -y, --yes                    Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud managed-kubernetes](ovhcloud_cloud_managed-kubernetes.md)	 - Manage Kubernetes clusters in the given cloud project

