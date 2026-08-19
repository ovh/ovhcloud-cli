## ovhcloud ip byoip

Aggregate or slice a bring-your-own-IP block

### Options

```
  -h, --help   help for byoip
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

* [ovhcloud ip](ovhcloud_ip.md)	 - Retrieve information and manage your IP services
* [ovhcloud ip byoip aggregate](ovhcloud_ip_byoip_aggregate.md)	 - Merge this block with its neighbours
* [ovhcloud ip byoip aggregations](ovhcloud_ip_byoip_aggregations.md)	 - List the blocks this one could be merged into
* [ovhcloud ip byoip slice](ovhcloud_ip_byoip_slice.md)	 - Split this block into smaller ones
* [ovhcloud ip byoip slices](ovhcloud_ip_byoip_slices.md)	 - List the ways this block could be split

