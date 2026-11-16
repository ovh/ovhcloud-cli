## ovhcloud cloud managed-kubernetes create

Create a new Kubernetes cluster

### Synopsis

Use this command to create a managed Kubernetes cluster in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-kubernetes create --name MyNewCluster --region SBG5 --plan FREE

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
      --api-server-admission-plugins-disabled strings         Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)
      --api-server-admission-plugins-enabled strings          Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)
      --api-server-authorized-ips strings                     CIDR blocks allowed to reach the API server, '0.0.0.0/0' to allow all (replaces the current allowlist when editing)
      --auto-upgrade-patch-enabled                            Automatically upgrade the cluster to the latest patch version (default true)
      --cilium-cluster-id uint8                               Cilium cluster ID
      --cilium-cluster-mesh-apiserver-node-port uint16        ClusterMesh API server node port
      --cilium-cluster-mesh-apiserver-service-type string     ClusterMesh API server service type (LoadBalancer, NodePort)
      --cilium-cluster-mesh-enabled                           Enable Cilium ClusterMesh
      --cilium-hubble-enabled                                 Enable Hubble observability
      --cilium-hubble-relay-enabled                           Enable Hubble Relay
      --cilium-hubble-ui-backend-limits-cpu string            Hubble UI backend CPU limit (e.g. '500m')
      --cilium-hubble-ui-backend-limits-memory string         Hubble UI backend memory limit (e.g. '256Mi')
      --cilium-hubble-ui-backend-requests-cpu string          Hubble UI backend CPU request (e.g. '100m')
      --cilium-hubble-ui-backend-requests-memory string       Hubble UI backend memory request (e.g. '128Mi')
      --cilium-hubble-ui-enabled                              Enable Hubble UI
      --cilium-hubble-ui-frontend-limits-cpu string           Hubble UI frontend CPU limit (e.g. '500m')
      --cilium-hubble-ui-frontend-limits-memory string        Hubble UI frontend memory limit (e.g. '256Mi')
      --cilium-hubble-ui-frontend-requests-cpu string         Hubble UI frontend CPU request (e.g. '100m')
      --cilium-hubble-ui-frontend-requests-memory string      Hubble UI frontend memory request (e.g. '128Mi')
      --cni-type string                                       CNI plugin of the cluster (CANAL for the FREE plan, CILIUM for the STANDARD plan). Derived from --plan when omitted
      --editor                                                Use a text editor to define parameters
      --from-file string                                      File containing parameters
  -h, --help                                                  help for create
      --init-file string                                      Create a file with example parameters
      --ip-allocation-policy-pods string                      IPv4 CIDR block allocated for pod networking (e.g. '10.0.0.0/16')
      --ip-allocation-policy-services string                  IPv4 CIDR block allocated for service networking (e.g. '10.96.0.0/12')
      --kube-proxy-iptables-min-sync-period string            Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-iptables-sync-period string                Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-min-sync-period string                Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-scheduler string                      Scheduler for kube-proxy ipvs (DESTINATION_HASHING, LEAST_CONNECTION, NEVER_QUEUE, ROUND_ROBIN, SHORTEST_EXPECTED_DELAY, SOURCE_HASHING)
      --kube-proxy-ipvs-sync-period string                    Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-tcp-fin-timeout string                Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-tcp-timeout string                    Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-udp-timeout string                    Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-mode string                                Kube-proxy mode (IPTABLES, IPVS), immutable after creation
      --name string                                           Name of the Kubernetes cluster (auto-generated at creation if omitted)
      --network-type string                                   Network topology of the cluster nodes (PUBLIC, VRACK). When omitted, VRACK if a vRack private network is given, PUBLIC otherwise
      --network-vrack-default-egress-type string              Default egress route of the cluster nodes (MANAGED_GATEWAY, NODE_PUBLIC_INTERFACE, USER_GATEWAY). Defaults to MANAGED_GATEWAY on STANDARD plan, NODE_PUBLIC_INTERFACE on FREE plan
      --network-vrack-default-egress-user-gateway-ip string   IP address of the egress gateway, required when the default egress type is USER_GATEWAY
      --network-vrack-load-balancer-subnet-id string          Subnet ID used by the load balancers (VRACK topology)
      --network-vrack-private-network-id string               Private network ID used by the cluster nodes (VRACK topology)
      --network-vrack-subnet-id string                        Subnet ID used by the cluster nodes (VRACK topology)
      --oidc-ca-content string                                CA certificate content for the OIDC provider
      --oidc-client-id string                                 OIDC client ID
      --oidc-groups-claim strings                             OIDC groups claim(s)
      --oidc-groups-prefix string                             Prefix prepended to group claims
      --oidc-issuer-url string                                OIDC issuer URL
      --oidc-required-claim strings                           OIDC required claim(s)
      --oidc-signing-algs strings                             OIDC signing algorithm(s) (ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512)
      --oidc-username-claim string                            OIDC username claim
      --oidc-username-prefix string                           Prefix prepended to username claims
      --plan string                                           Plan of the Kubernetes cluster: FREE, STANDARD (required)
      --region string                                         Region for the Kubernetes cluster (required)
      --replace                                               Replace parameters file if it already exists
      --version string                                        Kubernetes version (a minor version such as '1.34' is only accepted at creation time, which defaults to the latest supported version if omitted)
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

* [ovhcloud cloud managed-kubernetes](ovhcloud_cloud_managed-kubernetes.md)	 - Manage Kubernetes clusters in the given cloud project

