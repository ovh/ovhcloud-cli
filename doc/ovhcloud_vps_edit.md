## ovhcloud vps edit

Edit the given VPS

```
ovhcloud vps edit <service_name> [flags]
```

### Options

```
      --display-name string   Display name of the VPS
      --editor                Use a text editor to define parameters
  -h, --help                  help for edit
      --keymap string         Keymap of the VPS (fr, us)
      --netboot-mode string   Netboot mode of the VPS (local, rescue)
      --sla-monitoring        Enable or disable SLA monitoring for the VPS
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud vps](ovhcloud_vps.md)	 - Retrieve information and manage your VPS services

