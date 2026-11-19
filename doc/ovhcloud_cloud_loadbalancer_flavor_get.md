## ovhcloud cloud loadbalancer flavor get

Get details of a specific loadbalancer flavor, optionally restricted to one region

### Synopsis

Get details of a specific loadbalancer flavor, using its name (SMALL, MEDIUM, ...).
Without a region, a flavor sold in several regions is listed once per region.

```
ovhcloud cloud loadbalancer flavor get [region] <flavor_name> [flags]
```

### Options

```
  -h, --help   help for get
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

* [ovhcloud cloud loadbalancer flavor](ovhcloud_cloud_loadbalancer_flavor.md)	 - List available loadbalancer flavors in the given cloud project

