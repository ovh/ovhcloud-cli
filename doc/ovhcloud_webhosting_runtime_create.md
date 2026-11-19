## ovhcloud webhosting runtime create

Create a runtime

```
ovhcloud webhosting runtime create <service_name> [flags]
```

### Options

```
      --app-bootstrap string   Application bootstrap script
      --app-env string         Application environment
      --domain strings         Domains to attach
      --editor                 Use a text editor to define parameters
      --from-file string       File containing parameters
  -h, --help                   help for create
      --name string            Runtime name
      --public-dir string      Public directory
      --runtime-default        Set as default runtime
      --type string            Runtime backend type
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting runtime](ovhcloud_webhosting_runtime.md)	 - Manage runtimes

