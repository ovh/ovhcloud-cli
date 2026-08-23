## ovhcloud webhosting token

Get DNS verification token

### Synopsis

Use to link an external domain. This token must be added to a TXT record on your DNS zone using the ovhcontrol subdomain.

```
ovhcloud webhosting token <service_name> [flags]
```

### Options

```
  -h, --help   help for token
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

