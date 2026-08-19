## ovhcloud cloud loadbalancer listener edit

Edit a specific listener

```
ovhcloud cloud loadbalancer listener edit <loadbalancer_id> <listener_id> [flags]
```

### Options

```
      --default-pool-id string             Default pool ID
      --default-tls-container-ref string   Reference of the default TLS certificate container
      --description string                 Description of the listener
      --editor                             Use a text editor to define parameters
  -h, --help                               help for edit
      --name string                        Name of the listener
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

