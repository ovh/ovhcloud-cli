## ovhcloud baremetal boot

Manage boot options for the given baremetal

### Options

```
  -h, --help   help for boot
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
```

### SEE ALSO

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services
* [ovhcloud baremetal boot list](ovhcloud_baremetal_boot_list.md)	 - List boot options for the given baremetal
* [ovhcloud baremetal boot set](ovhcloud_baremetal_boot_set.md)	 - Configure a boot ID on the given baremetal
* [ovhcloud baremetal boot set-script](ovhcloud_baremetal_boot_set-script.md)	 - Configure a boot script on the given baremetal

