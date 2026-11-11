## ovhcloud cloud loadbalancer edit

Edit the given loadbalancer

```
ovhcloud cloud loadbalancer edit <loadbalancer_id> [flags]
```

### Options

```
      --description string   Description of the loadbalancer
      --editor               Use a text editor to define parameters
  -h, --help                 help for edit
      --name string          Name of the loadbalancer
      --size string          Size of the loadbalancer (e.g. small, medium, large) or flavor UUID
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

