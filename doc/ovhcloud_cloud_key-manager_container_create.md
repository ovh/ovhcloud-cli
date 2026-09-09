## ovhcloud cloud key-manager container create

Create a new Key Manager container

```
ovhcloud cloud key-manager container create [flags]
```

### Options

```
      --availability-zone string   Availability zone within the region
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Desired container name
      --region string              Region code where the container is located
      --replace                    Replace parameters file if it already exists
      --secret-ref stringArray     Secret reference as '<name>=<secretId>' (repeatable)
      --type string                Type of the container (CERTIFICATE, GENERIC, RSA)
      --wait                       Wait for the container to be ready before exiting
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

* [ovhcloud cloud key-manager container](ovhcloud_cloud_key-manager_container.md)	 - Manage Key Manager containers

