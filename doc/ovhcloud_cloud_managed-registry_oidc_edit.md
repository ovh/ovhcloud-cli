## ovhcloud cloud managed-registry oidc edit

Edit the OIDC configuration for a container registry

```
ovhcloud cloud managed-registry oidc edit <registry_id> [flags]
```

### Options

```
      --admin-group string     Group granted admin role
      --auto-onboard           Automatically create users on first login
      --client-id string       OIDC client ID
      --client-secret string   OIDC client secret
      --editor                 Use a text editor to define parameters
      --endpoint string        OIDC provider endpoint
      --group-filter string    Regex applied to filter groups
      --groups-claim string    OIDC claim containing groups
  -h, --help                   help for edit
      --name string            OIDC provider name
      --scope string           OIDC scopes
      --user-claim string      OIDC claim containing the username
      --verify-cert            Verify the provider TLS certificate
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

* [ovhcloud cloud managed-registry oidc](ovhcloud_cloud_managed-registry_oidc.md)	 - Manage container registry OIDC integration

