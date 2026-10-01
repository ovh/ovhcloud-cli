## ovhcloud cloud instance application-access

Get application access credentials for the given instance

### Synopsis

Get the credentials to access the application installed on the given instance (e.g. WordPress, GitLab, etc.)

```
ovhcloud cloud instance application-access <instance_id> [flags]
```

### Options

```
  -h, --help   help for application-access
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

