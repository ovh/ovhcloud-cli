## ovhcloud cloud managed-kubernetes reset

Reset the given Kubernetes cluster

### Synopsis

Reset the given Kubernetes cluster.
All Kubernetes data will be erased (pods, services, configuration, etc), nodes will be either deleted or reinstalled.

There are three ways to define the reset parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-kubernetes reset <cluster_id> --name MyNewCluster --version 1.32

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud managed-kubernetes reset <cluster_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct parameters, run:

	ovhcloud cloud managed-kubernetes reset <cluster_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud managed-kubernetes reset <cluster_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud managed-kubernetes reset <cluster_id> --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud managed-kubernetes reset <cluster_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reset will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-kubernetes reset <cluster_id> --editor --version 1.31


```
ovhcloud cloud managed-kubernetes reset <cluster_id> [flags]
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
  -h, --help                                                help for reset
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
      --private-network-id string                           OpenStack private network ID that the cluster will use
      --private-network.default-vrack-gateway string        If defined, all egress traffic will be routed towards this IP address, which should belong to the private network
      --private-network.routing-as-default                  Set private network routing as default
      --replace                                             Replace parameters file if it already exists
      --update-policy string                                Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)
      --version string                                      Kubernetes version
      --worker-nodes-policy string                          Worker nodes reset policy (delete, reinstall)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-kubernetes](ovhcloud_cloud_managed-kubernetes.md)	 - Manage Kubernetes clusters in the given cloud project

