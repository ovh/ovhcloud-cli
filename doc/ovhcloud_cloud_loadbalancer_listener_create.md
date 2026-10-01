## ovhcloud cloud loadbalancer listener create

Create a listener in the given region

```
ovhcloud cloud loadbalancer listener create <region> [flags]
```

### Options

```
      --editor                   Use a text editor to define parameters
      --from-file string         File containing parameters
  -h, --help                     help for create
      --init-file string         Create a file with example parameters
      --loadbalancer-id string   Loadbalancer ID
      --name string              Name of the listener
      --port int                 Port to listen on
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

* [ovhcloud cloud loadbalancer listener](ovhcloud_cloud_loadbalancer_listener.md)	 - Manage listeners of loadbalancers

