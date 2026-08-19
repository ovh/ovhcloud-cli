## ovhcloud baremetal vni ola-create-aggregation

Group interfaces into an aggregation

```
ovhcloud baremetal vni ola-create-aggregation <service_name> --name <name> --interface <uuid> --interface <uuid> [flags]
```

### Options

```
  -h, --help                    help for ola-create-aggregation
      --interface stringArray   Interfaces to group
      --name string             Name of the aggregation
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud baremetal vni](ovhcloud_baremetal_vni.md)	 - Manage Virtual Network Interfaces of the given baremetal

