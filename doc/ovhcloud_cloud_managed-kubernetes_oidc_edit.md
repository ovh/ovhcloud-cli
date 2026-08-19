## ovhcloud cloud managed-kubernetes oidc edit

Edit the OIDC configuration for the given Kubernetes cluster

```
ovhcloud cloud managed-kubernetes oidc edit <cluster_id> [flags]
```

### Options

```
      --ca-content string            CA certificate content for the OIDC provider
      --client-id string             OIDC client ID
      --groups-claim strings         OIDC groups claim(s)
      --groups-prefix string         Prefix prepended to group claims
  -h, --help                         help for edit
      --issuer-url string            OIDC issuer URL
      --required-claim strings       OIDC required claim(s)
      --signing-algorithms strings   OIDC signing algorithm(s) (ES256, ES384, ES512, PS256, PS384, PS512, RS256, RS384, RS512)
      --username-claim string        OIDC username claim
      --username-prefix string       Prefix prepended to username claims
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

* [ovhcloud cloud managed-kubernetes oidc](ovhcloud_cloud_managed-kubernetes_oidc.md)	 - Manage OpenID Connect (OIDC) integration for Kubernetes clusters

