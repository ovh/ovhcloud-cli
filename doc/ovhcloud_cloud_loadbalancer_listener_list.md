## ovhcloud cloud loadbalancer listener list

List all listeners

```
ovhcloud cloud loadbalancer listener list [flags]
```

### Options

```
      --filter stringArray       Filter results by any property using https://github.com/PaesslerAG/gval syntax
                                 Examples:
                                   --filter 'state=="running"'
                                   --filter 'name=~"^my.*"'
                                   --filter 'nested.property.subproperty>10'
                                   --filter 'startDate>="2023-12-01"'
                                   --filter 'name=~"something" && nbField>10'
  -h, --help                     help for list
      --loadbalancer-id string   Filter listeners by loadbalancer ID
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

