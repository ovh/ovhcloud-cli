## ovhcloud baremetal power

Power the given baremetal off and on

### Synopsis

Power the given dedicated server off and on.

There is no power switch in the API: powering off works by setting the server's
"Power-off server" boot entry and rebooting into it. That has a consequence
worth knowing — a server left on that entry shuts itself down at every reboot,
including a reboot asked for from the manager. "power on" therefore puts the
previous boot back before starting the server, and "power status" says so when
a server is sitting on the power-off entry.

### Options

```
  -h, --help   help for power
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
                         Examples:
                           --output json
                           --output yaml
                           --output interactive
                           --output 'id' (to extract a single field)
                           --output 'nested.field.subfield' (to extract a nested field)
                           --output '[id, "name"]' (to extract multiple fields as an array)
                           --output '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                           --output 'name+","+type' (to extract and concatenate fields in a string)
                           --output '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services
* [ovhcloud baremetal power off](ovhcloud_baremetal_power_off.md)	 - Power off the given baremetal
* [ovhcloud baremetal power on](ovhcloud_baremetal_power_on.md)	 - Power on the given baremetal, restoring the boot it was on
* [ovhcloud baremetal power status](ovhcloud_baremetal_power_status.md)	 - Show whether the server is on, and what it will boot on

