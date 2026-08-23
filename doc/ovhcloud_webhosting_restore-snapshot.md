## ovhcloud webhosting restore-snapshot

Restore a snapshot

```
ovhcloud webhosting restore-snapshot <service_name> [flags]
```

### Options

```
      --backup string      Backup to restore (allowed: daily.1, daily.2, daily.3, weekly.1, weekly.2)
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for restore-snapshot
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services

