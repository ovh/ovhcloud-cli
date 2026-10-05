## ovhcloud cloud instance backup create

Create a backup of the given instance

```
ovhcloud cloud instance backup create <instance_id> <backup_name> [flags]
```

### Options

```
      --distant-backup-name string   Name of the backup in the distant region (for cross region backup)
      --distant-region-name string   Name of the distant region (for cross region backup)
  -h, --help                         help for create
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

* [ovhcloud cloud instance backup](ovhcloud_cloud_instance_backup.md)	 - Manage backups of the given instance

