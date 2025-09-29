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
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services

