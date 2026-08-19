## ovhcloud baremetal edit

Update the given baremetal

```
ovhcloud baremetal edit <service_name> [flags]
```

### Options

```
      --boot-id int                  Boot ID
      --boot-script string           Boot script
      --editor                       Use a text editor to define parameters
      --efi-bootloader-path string   EFI bootloader path
  -h, --help                         help for edit
      --monitoring                   Enable monitoring
      --no-intervention              Disable interventions
      --rescue-mail string           Rescue mail
      --rescue-ssh-key string        Rescue SSH key
      --root-device string           Root device
      --state string                 State (e.g., error)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services

