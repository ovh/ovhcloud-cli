## ovhcloud cloud instance autobackup create

Create an automatic backup workflow for the given instance

```
ovhcloud cloud instance autobackup create <instance_id> [flags]
```

### Options

```
      --cron string    Unix Cron pattern (e.g. '0 0 * * *')
  -h, --help           help for create
      --name string    Workflow name
      --rotation int   Number of backups to keep
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

* [ovhcloud cloud instance autobackup](ovhcloud_cloud_instance_autobackup.md)	 - Manage automatic backup workflows for instances

