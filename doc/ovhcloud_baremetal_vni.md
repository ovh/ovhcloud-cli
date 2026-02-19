## ovhcloud baremetal vni

Manage Virtual Network Interfaces of the given baremetal

### Options

```
  -h, --help   help for vni
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
* [ovhcloud baremetal vni list](ovhcloud_baremetal_vni_list.md)	 - List Virtual Network Interfaces of the given baremetal
* [ovhcloud baremetal vni ola-create-aggregation](ovhcloud_baremetal_vni_ola-create-aggregation.md)	 - Group interfaces into an aggregation
* [ovhcloud baremetal vni ola-reset](ovhcloud_baremetal_vni_ola-reset.md)	 - Reset interfaces to default configuration

