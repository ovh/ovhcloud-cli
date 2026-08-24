## ovhcloud ip game rule

Manage game anti-DDoS rules

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
```

### SEE ALSO

* [ovhcloud ip game](ovhcloud_ip_game.md)	 - Manage the game anti-DDoS filter on the given IP block
* [ovhcloud ip game rule add](ovhcloud_ip_game_rule_add.md)	 - Open a port range for one game protocol
* [ovhcloud ip game rule delete](ovhcloud_ip_game_rule_delete.md)	 - Delete a game anti-DDoS rule
* [ovhcloud ip game rule get](ovhcloud_ip_game_rule_get.md)	 - Show one game anti-DDoS rule
* [ovhcloud ip game rule list](ovhcloud_ip_game_rule_list.md)	 - List the game anti-DDoS rules of an address

