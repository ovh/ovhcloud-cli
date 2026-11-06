## ovhcloud cloud key-manager secret consumer register

Register a consumer for the given secret

```
ovhcloud cloud key-manager secret consumer register <secret_id> [flags]
```

### Options

```
  -h, --help                   help for register
      --resource-id string     UUID of the resource consuming the secret/container
      --resource-type string   Type of the consuming resource (IMAGE, INSTANCE, LOADBALANCER)
      --service string         OpenStack service type of the consumer (COMPUTE, IMAGE, LOADBALANCER, NETWORK)
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

* [ovhcloud cloud key-manager secret consumer](ovhcloud_cloud_key-manager_secret_consumer.md)	 - Manage consumers of a Key Manager secret

