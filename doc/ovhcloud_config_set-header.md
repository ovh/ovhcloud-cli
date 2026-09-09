## ovhcloud config set-header

Set a custom HTTP header to send on every API request

```
ovhcloud config set-header <name> <value>
```

### Examples

```
ovhcloud config set-header X-Routing-Key abc123
```

### Options

```
  -h, --help   help for set-header
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud config](ovhcloud_config.md)	 - Manage your CLI configuration

