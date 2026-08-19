## ovhcloud cloud managed-registry create

Create a new container registry

```
ovhcloud cloud managed-registry create [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for create
      --init-file string   Create a file with example parameters
      --name string        Name of the container registry
      --plan-id string     Plan ID for the container registry. Available plans can be listed with 'ovhcloud cloud reference managed-registry list-plans'
      --region string      Region for the container registry (e.g., DE, GRA, BHS)
      --replace            Replace parameters file if it already exists
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

* [ovhcloud cloud managed-registry](ovhcloud_cloud_managed-registry.md)	 - Manage container registries in the given cloud project

