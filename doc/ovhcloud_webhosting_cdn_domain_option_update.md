## ovhcloud webhosting cdn domain option update

Update a CDN domain option

```
ovhcloud webhosting cdn domain option update <service_name> <domain> <option> [flags]
```

### Options

```
      --destination string        Destination URL for redirects
      --editor                    Use a text editor to define parameters
      --enabled                   Enable or disable the option
      --follow-uri                Follow URI on redirects
      --from-file string          File containing parameters
  -h, --help                      help for update
      --origins string            Authorized origins (comma separated)
      --pattern string            URL pattern for the option
      --pattern-type string       Pattern type
      --priority int              Cache rule priority
      --query-parameters string   Action on query parameters
      --resource strings          Resource URI (repeatable)
      --status-code int           Redirection HTTP status code
      --ttl int                   Cache time in seconds
      --type string               Option type
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting cdn domain option](ovhcloud_webhosting_cdn_domain_option.md)	 - Manage CDN domain options

