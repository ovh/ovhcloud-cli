## ovhcloud baremetal vrack

Attach the given baremetal to a vRack, or detach it

### Options

```
  -h, --help   help for vrack
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
* [ovhcloud baremetal vrack attach](ovhcloud_baremetal_vrack_attach.md)	 - Attach the given baremetal to a vRack
* [ovhcloud baremetal vrack detach](ovhcloud_baremetal_vrack_detach.md)	 - Detach the given baremetal from its vRack
* [ovhcloud baremetal vrack show](ovhcloud_baremetal_vrack_show.md)	 - Show the vRack the given baremetal is in

