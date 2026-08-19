## ovhcloud webhosting user update

Update a FTP/SSH user

```
ovhcloud webhosting user update <service_name> <login> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
  -h, --help               help for update
      --home string        Home directory for the FTP/SSH user
      --ssh-state string   SSH state (allowed: active, none, sftponly)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting user](ovhcloud_webhosting_user.md)	 - Manage FTP/SSH users

