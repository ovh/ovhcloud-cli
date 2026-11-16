## ovhcloud cloud loadbalancer pool create

Create a pool on a loadbalancer

```
ovhcloud cloud loadbalancer pool create <loadbalancer_id> [flags]
```

### Options

```
      --algorithm string                       Algorithm (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP)
      --description string                     Description of the pool
      --editor                                 Use a text editor to define parameters
      --from-file string                       File containing parameters
      --health-monitor-delay int               Duration between sending probes to members, in seconds
      --health-monitor-expected-codes string   Expected HTTP status codes of HTTP and HTTPS health monitors (e.g. 200, 200-204)
      --health-monitor-http-method string      HTTP method used by HTTP and HTTPS health monitors (GET, HEAD, POST, ...)
      --health-monitor-max-retries int         Number of successful checks before changing status to ONLINE
      --health-monitor-max-retries-down int    Number of allowed check failures before changing status to ERROR
      --health-monitor-timeout int             Maximum time in seconds to connect before timeout
      --health-monitor-type string             Type of the health monitor (HTTP, HTTPS, PING, TCP, TLS_HELLO, UDP_CONNECT, SCTP)
      --health-monitor-url-path string         URL path requested by HTTP and HTTPS health monitors
  -h, --help                                   help for create
      --init-file string                       Create a file with example parameters
      --name string                            Name of the pool
      --protocol string                        Protocol (HTTP, HTTPS, PROXY, PROXYV2, TCP, UDP, SCTP)
      --replace                                Replace parameters file if it already exists
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

* [ovhcloud cloud loadbalancer pool](ovhcloud_cloud_loadbalancer_pool.md)	 - Manage pools of loadbalancers

