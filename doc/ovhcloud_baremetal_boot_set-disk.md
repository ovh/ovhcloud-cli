## ovhcloud baremetal boot set-disk

Configure the given baremetal to boot on its hard disk

### Synopsis

Restore the hard disk boot entry of the given dedicated server.

This is the counterpart of "baremetal reboot-rescue": a server left with the
rescue boot entry will come back in rescue mode at its next reboot, whatever
triggers it. The change applies at the next reboot:

	ovhcloud baremetal boot set-disk ns1234.ip-11.22.33.net
	ovhcloud baremetal reboot ns1234.ip-11.22.33.net


```
ovhcloud baremetal boot set-disk <service_name> [flags]
```

### Options

```
  -h, --help   help for set-disk
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

* [ovhcloud baremetal boot](ovhcloud_baremetal_boot.md)	 - Manage boot options for the given baremetal

