## ovhcloud ip game

Manage the game anti-DDoS filter on the given IP block

### Options

```
  -h, --help   help for game
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
* [ovhcloud ip game edit](ovhcloud_ip_game_edit.md)	 - Turn UDP firewall mode on or off
* [ovhcloud ip game get](ovhcloud_ip_game_get.md)	 - Show the game anti-DDoS configuration of an address
* [ovhcloud ip game list](ovhcloud_ip_game_list.md)	 - List the addresses under game anti-DDoS
* [ovhcloud ip game rule](ovhcloud_ip_game_rule.md)	 - Manage game anti-DDoS rules

