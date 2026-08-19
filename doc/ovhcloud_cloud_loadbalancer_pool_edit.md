## ovhcloud cloud loadbalancer pool edit

Edit a specific pool

### Synopsis

Edit a specific pool, including the health monitor nested in it.
To remove the health monitor, use --editor and delete the "healthMonitor" block.

```
ovhcloud cloud loadbalancer pool edit <loadbalancer_id> <pool_id> [flags]
```

### Options

```
      --algorithm string                       Algorithm (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP)
      --description string                     Description of the pool
      --editor                                 Use a text editor to define parameters
      --health-monitor-delay int               Time in seconds between sending health check probes to members
      --health-monitor-expected-codes string   Expected HTTP status codes of HTTP and HTTPS health monitors (e.g. 200, 200-204)
      --health-monitor-http-method string      HTTP method used by HTTP and HTTPS health monitors (GET, HEAD, POST, ...)
      --health-monitor-max-retries int         Number of consecutive health check failures before marking a member as INACTIVE (1-10)
      --health-monitor-max-retries-down int    Number of consecutive health check failures before marking a member as ERROR (1-10)
      --health-monitor-timeout int             Maximum time in seconds to wait for a health check reply (must be less than or equal to the delay)
      --health-monitor-type string             Type of the health monitor (HTTP, HTTPS, PING, TCP, TLS_HELLO, UDP_CONNECT, SCTP)
      --health-monitor-url-path string         URL path requested by HTTP and HTTPS health monitors
  -h, --help                                   help for edit
      --name string                            Name of the pool
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

