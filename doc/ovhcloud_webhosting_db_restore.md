## ovhcloud webhosting db restore

Restore database from snapshot date

```
ovhcloud webhosting db restore <service_name> <name> [flags]
```

### Options

```
      --date string        Dump type to restore (allowed: daily.1, now, weekly.1)
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for restore
      --send-email         Send email when restore completes
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases

