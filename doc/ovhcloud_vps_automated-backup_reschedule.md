## ovhcloud vps automated-backup reschedule

Reschedule the automated backup of the given VPS

```
ovhcloud vps automated-backup reschedule <service_name> <time> [flags]
```

### Examples

```
ovh-cli vps automated-backup reschedule my-vps 15:04:05
```

### Options

```
  -h, --help   help for reschedule
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using gval format)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud vps automated-backup](ovhcloud_vps_automated-backup.md)	 - Manage VPS automated backups

