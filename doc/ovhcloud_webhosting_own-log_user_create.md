## ovhcloud webhosting own-log user create

Create an own log user

```
ovhcloud webhosting own-log user create <service_name> <ownlog_id> [flags]
```

### Options

```
      --description string   Description for this user (required)
      --editor               Use a text editor to define parameters
      --from-file string     File containing parameters
  -h, --help                 help for create
      --login string         User login used to connect to logs.ovh.net
      --password string      User password (required)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting own-log user](ovhcloud_webhosting_own-log_user.md)	 - Manage own log users

