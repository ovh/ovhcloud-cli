## ovhcloud cloud loadbalancer associate-floating-ip

Associate an existing floating IP to a loadbalancer

```
ovhcloud cloud loadbalancer associate-floating-ip <loadbalancer_id> [flags]
```

### Options

```
      --editor                  Use a text editor to define parameters
      --floating-ip-id string   Floating IP ID
      --from-file string        File containing parameters
  -h, --help                    help for associate-floating-ip
      --init-file string        Create a file with example parameters
      --ip string               Private loadbalancer IP to associate the floating IP with
      --replace                 Replace parameters file if it already exists
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

* [ovhcloud cloud loadbalancer](ovhcloud_cloud_loadbalancer.md)	 - Manage loadbalancers in the given cloud project

