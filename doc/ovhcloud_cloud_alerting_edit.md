## ovhcloud cloud alerting edit

Edit a billing alert configuration

```
ovhcloud cloud alerting edit <alert_id> [flags]
```

### Options

```
      --delay int               Delay between alerts in seconds (minimum 3600)
      --editor                  Use a text editor to define parameters
      --emails strings          Email addresses to receive alerts (comma-separated)
  -h, --help                    help for edit
      --monthly-threshold int   Monthly threshold value
      --name string             Alert name
      --service string          Service of the alert. Allowed: ai_endpoint, all, block_storage, data_platform, instances, instances_gpu, instances_without_gpu, objet_storage, rancher, snapshot
      --status string           Status of the alert. Allowed: deleted, disabled, ok
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

