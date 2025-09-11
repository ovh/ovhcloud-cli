## ovhcloud cloud kube create

Create a new Kubernetes cluster

### Synopsis

Use this command to create a managed Kubernetes cluster in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

  ovhcloud cloud kube create --name MyNewCluster --region SBG5 --version 1.32 …

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud kube create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud kube create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud kube create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud kube create --from-file ./params.json --name NameOverriden

3. Using your default text editor:

  ovhcloud cloud kube create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud kube create --editor --region BHS5


```
ovhcloud cloud kube create [flags]
```

### Options

```
      --customization.api-server.admission-plugins.disabled strings   Admission plugins to disable on API server (AlwaysPullImages, NodeRestriction)
      --customization.api-server.admission-plugins.enabled strings    Admission plugins to enable on API server (AlwaysPullImages, NodeRestriction)
      --customization.kube-proxy.iptables.min-sync-period string      Minimum period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --customization.kube-proxy.iptables.sync-period string          Period that iptables rules are refreshed, in RFC3339 duration format (e.g. 'PT60S')
      --customization.kube-proxy.ipvs.min-sync-period string          Minimum period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --customization.kube-proxy.ipvs.scheduler string                Scheduler for kube-proxy ipvs (dh, lc, nq, rr, sed, sh)
      --customization.kube-proxy.ipvs.sync-period string              Period that ipvs rules are refreshed in RFC3339 duration format (e.g. 'PT60S')
      --customization.kube-proxy.ipvs.tcp-fin-timeout string          Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format (e.g. 'PT60S')
      --customization.kube-proxy.ipvs.tcp-timeout string              Timeout value used for idle IPVS TCP sessions in RFC3339 duration format (e.g. 'PT60S')
      --customization.kube-proxy.ipvs.udp-timeout string              Timeout value used for IPVS UDP packets in RFC3339 duration format (e.g. 'PT60S')
      --editor                                                        Use a text editor to define parameters
      --from-file string                                              File containing parameters
  -h, --help                                                          help for create
      --init-file string                                              Create a file with example parameters
      --kube-proxy-mode string                                        Kube-proxy mode (iptables or ipvs)
      --load-balancers-subnet-id string                               OpenStack subnet ID that the load balancers will use
      --name string                                                   Name of the Kubernetes cluster
      --nodes-subnet-id string                                        OpenStack subnet ID that the cluster nodes will use
      --private-network-id string                                     OpenStack private network ID that the cluster will use
      --private-network.default-vrack-gateway string                  If defined, all egress traffic will be routed towards this IP address, which should belong to the private network
      --private-network.routing-as-default                            Set private network routing as default
      --region string                                                 Region for the Kubernetes cluster
      --replace                                                       Replace parameters file if it already exists
      --update-policy string                                          Update policy for the cluster (ALWAYS_UPDATE, MINIMAL_DOWNTIME, NEVER_UPDATE)
      --version string                                                Kubernetes version
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

* [ovhcloud cloud kube](ovhcloud_cloud_kube.md)	 - List Kubernetes clusters in the given cloud project

###### Auto generated by spf13/cobra on 26-Aug-2025
