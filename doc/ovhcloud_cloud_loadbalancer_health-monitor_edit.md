## ovhcloud cloud loadbalancer health-monitor edit

Edit a specific health monitor

```
ovhcloud cloud loadbalancer health-monitor edit <health_monitor_id> [flags]
```

### Options

```
      --delay int              Duration between sending probes to members, in seconds
      --editor                 Use a text editor to define parameters
  -h, --help                   help for edit
      --max-retries int        Number of successful checks before changing status to ONLINE
      --max-retries-down int   Number of allowed check failures before changing status to ERROR
      --name string            Name of the health monitor
      --timeout int            Maximum time in seconds to connect before timeout
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

* [ovhcloud cloud loadbalancer health-monitor](ovhcloud_cloud_loadbalancer_health-monitor.md)	 - Manage health monitors of loadbalancers

