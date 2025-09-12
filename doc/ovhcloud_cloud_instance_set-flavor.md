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
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project

