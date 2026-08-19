## ovhcloud baremetal boot set-script

Configure a boot script on the given baremetal

```
ovhcloud baremetal boot set-script <service_name> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for set-script
      --script string      Boot script to set on the baremetal
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud baremetal boot](ovhcloud_baremetal_boot.md)	 - Manage boot options for the given baremetal

