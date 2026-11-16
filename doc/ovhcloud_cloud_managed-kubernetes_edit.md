## ovhcloud cloud managed-kubernetes edit

Edit the given Kubernetes cluster

### Synopsis

Edit the given Kubernetes cluster. Only the mutable part of the cluster specification can be changed: to change the region, the plan, the network topology or the pods and services CIDRs, a new cluster must be created.

```
ovhcloud cloud managed-kubernetes edit <cluster_id> [flags]
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
      --editor                                                Use a text editor to define parameters
  -h, --help                                                  help for edit
      --kube-proxy-iptables-min-sync-period string            Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-iptables-sync-period string                Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-min-sync-period string                Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-scheduler string                      Scheduler for kube-proxy ipvs (DESTINATION_HASHING, LEAST_CONNECTION, NEVER_QUEUE, ROUND_ROBIN, SHORTEST_EXPECTED_DELAY, SOURCE_HASHING)
      --kube-proxy-ipvs-sync-period string                    Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-tcp-fin-timeout string                Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-tcp-timeout string                    Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy-ipvs-udp-timeout string                    Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')
      --name string                                           Name of the Kubernetes cluster (auto-generated at creation if omitted)
      --network-vrack-default-egress-type string              Default egress route of the cluster nodes (MANAGED_GATEWAY, NODE_PUBLIC_INTERFACE, USER_GATEWAY). Defaults to MANAGED_GATEWAY on STANDARD plan, NODE_PUBLIC_INTERFACE on FREE plan
      --network-vrack-default-egress-user-gateway-ip string   IP address of the egress gateway, required when the default egress type is USER_GATEWAY
      --oidc-ca-content string                                CA certificate content for the OIDC provider
      --oidc-client-id string                                 OIDC client ID
      --oidc-groups-claim strings                             OIDC groups claim(s)
      --oidc-groups-prefix string                             Prefix prepended to group claims
      --oidc-issuer-url string                                OIDC issuer URL
      --oidc-required-claim strings                           OIDC required claim(s)
      --oidc-signing-algs strings                             OIDC signing algorithm(s) (ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512)
      --oidc-username-claim string                            OIDC username claim
      --oidc-username-prefix string                           Prefix prepended to username claims
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

