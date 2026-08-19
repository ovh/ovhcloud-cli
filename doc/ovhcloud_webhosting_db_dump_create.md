## ovhcloud webhosting db dump create

Request a database dump

```
ovhcloud webhosting db dump create <service_name> <name> [flags]
```

### Options

```
      --date string        Dump type (allowed: daily.1, now, weekly.1)
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for create
      --send-email         Send email when dump is ready (default true)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting db dump](ovhcloud_webhosting_db_dump.md)	 - Manage database dumps

