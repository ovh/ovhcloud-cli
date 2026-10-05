## ovhcloud cloud loadbalancer log generate-url

Generate a temporary URL to retrieve logs

```
ovhcloud cloud loadbalancer log generate-url <loadbalancer_id> [flags]
```

### Options

```
  -h, --help          help for generate-url
      --kind string   Log kind (e.g., haproxy)
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

* [ovhcloud cloud loadbalancer log](ovhcloud_cloud_loadbalancer_log.md)	 - Manage loadbalancer logs

