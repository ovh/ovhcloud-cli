## ovhcloud webhosting cron update

Update a cron task

```
ovhcloud webhosting cron update <service_name> <id> [flags]
```

### Options

```
      --command string       Command to execute
      --description string   Description
      --editor               Use a text editor to define parameters
      --email string         Email for stderr
      --frequency string     Frequency (crontab format)
  -h, --help                 help for update
      --language string      Language
      --status string        Status (allowed: disabled, enabled, suspended)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting cron](ovhcloud_webhosting_cron.md)	 - Manage cron tasks

