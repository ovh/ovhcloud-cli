## ovhcloud cloud managed-registry ip-restrictions management delete

Delete a management IP restriction from a container registry

```
ovhcloud cloud managed-registry ip-restrictions management delete <registry_id> [flags]
```

### Options

```
  -h, --help              help for delete
      --ip-block string   IP block in CIDR notation to delete (e.g., 192.0.2.0/24)
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

* [ovhcloud cloud managed-registry ip-restrictions management](ovhcloud_cloud_managed-registry_ip-restrictions_management.md)	 - Manage IP restrictions for container registry Harbor UI and API access

