## ovhcloud cloud managed-kubernetes customization edit

Edit the customization of the given Kubernetes cluster

```
ovhcloud cloud managed-kubernetes customization edit <cluster_id> [flags]
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
  -h, --help                                                help for edit
      --kube-proxy.iptables.min-sync-period string          Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.iptables.sync-period string              Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.min-sync-period string              Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.scheduler string                    Scheduler for kube-proxy ipvs (dh, lc, nq, rr, sed, sh)
      --kube-proxy.ipvs.sync-period string                  Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.tcp-fin-timeout string              Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.tcp-timeout string                  Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')
      --kube-proxy.ipvs.udp-timeout string                  Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')
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
                               
                               When extracting a single scalar field, the value is printed without surrounding
                               quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-kubernetes customization](ovhcloud_cloud_managed-kubernetes_customization.md)	 - Manage Kubernetes cluster customizations

