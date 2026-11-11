## ovhcloud vps automated-backup reschedule

Reschedule the automated backup of the given VPS

```
ovhcloud vps automated-backup reschedule <service_name> <time> [flags]
```

### Examples

```
ovhcloud vps automated-backup reschedule my-vps 15:04:05
```

### Options

```
  -h, --help   help for reschedule
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud vps automated-backup](ovhcloud_vps_automated-backup.md)	 - Manage VPS automated backups

