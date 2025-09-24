## ovhcloud pack-xdsl

Retrieve information and manage your PackXDSL services

### Options

```
  -h, --help   help for pack-xdsl
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud pack-xdsl edit](ovhcloud_pack-xdsl_edit.md)	 - Edit the given PackXDSL
* [ovhcloud pack-xdsl get](ovhcloud_pack-xdsl_get.md)	 - Retrieve information of a specific PackXDSL
* [ovhcloud pack-xdsl list](ovhcloud_pack-xdsl_list.md)	 - List your PackXDSL services

