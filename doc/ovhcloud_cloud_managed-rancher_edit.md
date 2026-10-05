## ovhcloud cloud managed-rancher edit

Edit the given Rancher service

```
ovhcloud cloud managed-rancher edit <rancher_id> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
  -h, --help               help for edit
      --iam-auth-enabled   Allow Rancher to use identities managed by OVHcloud IAM (Identity and Access Management) to control access
      --name string        Name of the managed Rancher service
      --plan string        Plan of the managed Rancher service (OVHCLOUD_EDITION, STANDARD)
      --version string     Version of the managed Rancher service
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

* [ovhcloud cloud managed-rancher](ovhcloud_cloud_managed-rancher.md)	 - Manage Rancher services in the given cloud project

