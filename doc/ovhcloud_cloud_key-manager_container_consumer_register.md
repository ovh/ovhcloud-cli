## ovhcloud cloud key-manager container consumer register

Register a consumer for the given container

```
ovhcloud cloud key-manager container consumer register <container_id> [flags]
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

* [ovhcloud cloud key-manager container consumer](ovhcloud_cloud_key-manager_container_consumer.md)	 - Manage consumers of a Key Manager container

