## ovhcloud webhosting db copy restore

Restore a database copy

```
ovhcloud webhosting db copy restore <service_name> <name> [flags]
```

### Options

```
      --copy-id string     Copy ID to restore
      --editor             Use a text editor to define parameters
      --flush              Flush database before restore
      --from-file string   File containing parameters
  -h, --help               help for restore
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting db copy](ovhcloud_webhosting_db_copy.md)	 - Manage database copies

