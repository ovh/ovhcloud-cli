## ovhcloud cloud loadbalancer health-monitor create

Create a health monitor in the given region

```
ovhcloud cloud loadbalancer health-monitor create <region> [flags]
```

### Options

```
      --delay int              Duration between sending probes to members, in seconds
      --editor                 Use a text editor to define parameters
      --from-file string       File containing parameters
  -h, --help                   help for create
      --init-file string       Create a file with example parameters
      --max-retries int        Number of successful checks before changing status to ONLINE
      --max-retries-down int   Number of allowed check failures before changing status to ERROR
      --monitor-type string    Type of the monitor (http, https, ping, tcp, tls-hello, udp-connect, sctp)
      --name string            Name of the health monitor
      --pool-id string         Pool ID
      --replace                Replace parameters file if it already exists
      --timeout int            Maximum time in seconds to connect before timeout
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
  -y, --yes                    Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud loadbalancer health-monitor](ovhcloud_cloud_loadbalancer_health-monitor.md)	 - Manage health monitors of loadbalancers

