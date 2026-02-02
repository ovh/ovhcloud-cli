## ovhcloud cloud container-registry oidc edit

Edit the OIDC configuration for a container registry

```
ovhcloud cloud container-registry oidc edit <registry_id> [flags]
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
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud container-registry oidc](ovhcloud_cloud_container-registry_oidc.md)	 - Manage container registry OIDC integration

