## ovhcloud ip mitigation

Manage DDoS mitigation on the given IP block

### Options

```
  -h, --help   help for mitigation
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
* [ovhcloud ip mitigation add](ovhcloud_ip_mitigation_add.md)	 - Put an address on permanent mitigation
* [ovhcloud ip mitigation get](ovhcloud_ip_mitigation_get.md)	 - Show the mitigation state of an address
* [ovhcloud ip mitigation list](ovhcloud_ip_mitigation_list.md)	 - List the addresses under mitigation
* [ovhcloud ip mitigation remove](ovhcloud_ip_mitigation_remove.md)	 - Remove an address from mitigation
* [ovhcloud ip mitigation set](ovhcloud_ip_mitigation_set.md)	 - Turn permanent mitigation on or off for an address already known to the system

