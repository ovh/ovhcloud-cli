## ovhcloud config unset-header

Remove a previously configured custom HTTP header

```
ovhcloud config unset-header <name>
```

### Examples

```
ovhcloud config unset-header X-Routing-Key
```

### Options

```
  -h, --help   help for unset-header
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

