## ovhcloud cloud loadbalancer pool create

Create a pool on a loadbalancer

### Synopsis

Use this command to create a pool, optionally with its health monitor.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --editor

3. Using only CLI flags, for example a pool created with a full health monitor in one command:

	ovhcloud cloud loadbalancer pool create <loadbalancer_id> --name my-pool --algorithm ROUND_ROBIN --protocol HTTP \
		--health-monitor-type HTTP --health-monitor-delay 10 --health-monitor-timeout 5 --health-monitor-max-retries 3 \
		--health-monitor-url-path /healthz --health-monitor-expected-codes 200

  When any --health-monitor-* flag is used with flags only, type, delay, timeout and max-retries are all required.


```
ovhcloud cloud loadbalancer pool create <loadbalancer_id> [flags]
```

### Options

```
      --algorithm string                       Algorithm (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP)
      --description string                     Description of the pool
      --editor                                 Use a text editor to define parameters
      --from-file string                       File containing parameters
      --health-monitor-delay int               Time in seconds between sending health check probes to members
      --health-monitor-expected-codes string   Expected HTTP status codes of HTTP and HTTPS health monitors (e.g. 200, 200-204)
      --health-monitor-http-method string      HTTP method used by HTTP and HTTPS health monitors (GET, HEAD, POST, ...)
      --health-monitor-max-retries int         Number of consecutive health check failures before marking a member as INACTIVE (1-10)
      --health-monitor-max-retries-down int    Number of consecutive health check failures before marking a member as ERROR (1-10)
      --health-monitor-timeout int             Maximum time in seconds to wait for a health check reply (must be less than or equal to the delay)
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

