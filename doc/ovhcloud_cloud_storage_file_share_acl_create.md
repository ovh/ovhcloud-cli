## ovhcloud cloud storage file share acl create

Create an ACL for the given share

```
ovhcloud cloud storage file share acl create <share_id> [flags]
```

### Options

```
      --access-level string   Access level (ro, rw)
      --access-to string      Access target (IP address or CIDR)
  -h, --help                  help for create
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
      --region string          Region (skip region discovery if set)
```

### SEE ALSO

* [ovhcloud cloud storage file share acl](ovhcloud_cloud_storage_file_share_acl.md)	 - Manage share access control lists

