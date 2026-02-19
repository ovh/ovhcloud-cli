## ovhcloud dedicated-nasha

Retrieve information and manage your Dedicated NasHA services

### Options

```
  -h, --help   help for dedicated-nasha
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud dedicated-nasha edit](ovhcloud_dedicated-nasha_edit.md)	 - Edit the given Dedicated NasHA
* [ovhcloud dedicated-nasha get](ovhcloud_dedicated-nasha_get.md)	 - Retrieve information of a specific Dedicated NasHA
* [ovhcloud dedicated-nasha list](ovhcloud_dedicated-nasha_list.md)	 - List your Dedicated NasHA services

