## ovhcloud ip firewall rule

Manage firewall rules

### Options

```
  -h, --help   help for rule
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
  -y, --yes              Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud ip firewall](ovhcloud_ip_firewall.md)	 - Manage firewall (Edge Firewall) on the given IP
* [ovhcloud ip firewall rule create](ovhcloud_ip_firewall_rule_create.md)	 - Create a new firewall rule
* [ovhcloud ip firewall rule delete](ovhcloud_ip_firewall_rule_delete.md)	 - Delete a firewall rule
* [ovhcloud ip firewall rule get](ovhcloud_ip_firewall_rule_get.md)	 - Get a specific firewall rule
* [ovhcloud ip firewall rule list](ovhcloud_ip_firewall_rule_list.md)	 - List firewall rules for the given IP

