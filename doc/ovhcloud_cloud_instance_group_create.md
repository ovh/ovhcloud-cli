## ovhcloud cloud instance group create

Create an instance group

```
ovhcloud cloud instance group create <name> <region> [flags]
```

### Options

```
  -h, --help          help for create
  -t, --type string   Group type: affinity or anti-affinity (default is affinity) (default "affinity")
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

* [ovhcloud cloud instance group](ovhcloud_cloud_instance_group.md)	 - Manage instance groups

