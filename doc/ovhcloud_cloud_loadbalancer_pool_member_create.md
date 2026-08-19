## ovhcloud cloud loadbalancer pool member create

Create member(s) in a specific pool

```
ovhcloud cloud loadbalancer pool member create <pool_id> [flags]
```

### Options

```
      --address string      IP address of the member
      --editor              Use a text editor to define parameters
      --from-file string    File containing parameters
  -h, --help                help for create
      --init-file string    Create a file with example parameters
      --name string         Name of the member
      --protocol-port int   Protocol port number of the member
      --replace             Replace parameters file if it already exists
      --weight int          Weight of the member (1-256)
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

* [ovhcloud cloud loadbalancer pool member](ovhcloud_cloud_loadbalancer_pool_member.md)	 - Manage members of a loadbalancer pool

