## ovhcloud cloud instance set-flavor

Migrate the given instance to the specified flavor

```
ovhcloud cloud instance set-flavor <instance_id> <flavor_id> [flags]
```

### Options

```
      --flavor-selector   Use the interactive flavor selector
  -h, --help              help for set-flavor
      --wait              Wait for instance to run with the desired flavor before exiting
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project

