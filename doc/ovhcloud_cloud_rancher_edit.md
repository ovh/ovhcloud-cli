## ovhcloud cloud rancher edit

Edit the given Rancher service

```
ovhcloud cloud rancher edit <rancher_id> [flags]
```

### Options

```
      --editor                        Use a text editor to define parameters
  -h, --help                          help for edit
      --ip-restrictions stringArray   List of IP restrictions (expected format: '<cidrBlock>,<description>')
      --name string                   Name of the managed Rancher service
      --plan string                   Plan of the managed Rancher service (OVHCLOUD_EDITION, STANDARD)
      --version string                Version of the managed Rancher service
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

* [ovhcloud cloud rancher](ovhcloud_cloud_rancher.md)	 - Manage Rancher services in the given cloud project

