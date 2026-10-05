## ovhcloud cloud loadbalancer pool create

Create a pool in the given region

```
ovhcloud cloud loadbalancer pool create <region> [flags]
```

### Options

```
      --algorithm string         Algorithm (roundRobin, leastConnections, sourceIp)
      --editor                   Use a text editor to define parameters
      --from-file string         File containing parameters
  -h, --help                     help for create
      --init-file string         Create a file with example parameters
      --listener-id string       Listener ID
      --loadbalancer-id string   Loadbalancer ID
      --name string              Name of the pool
      --protocol string          Protocol (http, https, tcp, udp, sctp, prometheus)
      --replace                  Replace parameters file if it already exists
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

* [ovhcloud cloud loadbalancer pool](ovhcloud_cloud_loadbalancer_pool.md)	 - Manage pools of loadbalancers

