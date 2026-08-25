## ovhcloud ip delegation

Manage the reverse delegation of an IPv6 subnet

### Options

```
  -h, --help   help for delegation
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
* [ovhcloud ip delegation add](ovhcloud_ip_delegation_add.md)	 - Delegate the reverse of this subnet to a name server
* [ovhcloud ip delegation get](ovhcloud_ip_delegation_get.md)	 - Show one reverse delegation target
* [ovhcloud ip delegation list](ovhcloud_ip_delegation_list.md)	 - List the name servers the reverse is delegated to
* [ovhcloud ip delegation remove](ovhcloud_ip_delegation_remove.md)	 - Stop delegating the reverse of this subnet to a name server

