## ovhcloud webhosting cdn domain option add

Add a CDN domain option

```
ovhcloud webhosting cdn domain option add <service_name> <domain> [flags]
```

### Options

```
      --destination string        Destination URL for redirects
      --editor                    Use a text editor to define parameters
      --enabled                   Enable or disable the option
      --follow-uri                Follow URI on redirects
      --from-file string          File containing parameters
  -h, --help                      help for add
      --name string               Option name
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
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
                         Examples:
                           --output json
                           --output yaml
                           --output interactive
                           --output 'id' (to extract a single field)
                           --output 'nested.field.subfield' (to extract a nested field)
                           --output '[id, "name"]' (to extract multiple fields as an array)
                           --output '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                           --output 'name+","+type' (to extract and concatenate fields in a string)
                           --output '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting cdn domain option](ovhcloud_webhosting_cdn_domain_option.md)	 - Manage CDN domain options

