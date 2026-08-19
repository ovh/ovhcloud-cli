## ovhcloud cloud loadbalancer pool edit

Edit a specific pool

```
ovhcloud cloud loadbalancer pool edit <pool_id> [flags]
```

### Options

```
      --algorithm string   Algorithm (roundRobin, leastConnections, sourceIp)
      --editor             Use a text editor to define parameters
  -h, --help               help for edit
      --name string        Name of the pool
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

