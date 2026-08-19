## ovhcloud cloud alerting create

Create a new billing alert configuration

```
ovhcloud cloud alerting create [flags]
```

### Options

```
      --delay int               Delay between alerts in seconds (minimum 3600) (default 3600)
      --editor                  Use a text editor to define parameters
      --emails strings          Email addresses to receive alerts (comma-separated)
      --from-file string        File containing parameters
  -h, --help                    help for create
      --init-file string        Create a file with example parameters
      --monthly-threshold int   Monthly threshold value
      --name string             Alert name
      --replace                 Replace parameters file if it already exists
      --service string          Service of the alert. Allowed: ai_endpoint, all, block_storage, data_platform, instances, instances_gpu, instances_without_gpu, objet_storage, rancher, snapshot
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

* [ovhcloud cloud alerting](ovhcloud_cloud_alerting.md)	 - Manage billing alert configurations in the given cloud project

