## ovhcloud webhosting env create

Create an env var

```
ovhcloud webhosting env create <service_name> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for create
      --key string         Variable name
      --type string        Variable type (allowed: integer, password, string)
      --value string       Variable value
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting env](ovhcloud_webhosting_env.md)	 - Manage environment variables

