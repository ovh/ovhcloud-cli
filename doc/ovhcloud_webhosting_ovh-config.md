## ovhcloud webhosting ovh-config

Manage .ovhconfig settings

### Options

```
  -h, --help   help for ovh-config
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

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting ovh-config capabilities](ovhcloud_webhosting_ovh-config_capabilities.md)	 - List available versions and containers
* [ovhcloud webhosting ovh-config change](ovhcloud_webhosting_ovh-config_change.md)	 - Change a .ovhconfig entry
* [ovhcloud webhosting ovh-config get](ovhcloud_webhosting_ovh-config_get.md)	 - Get a .ovhconfig entry
* [ovhcloud webhosting ovh-config list](ovhcloud_webhosting_ovh-config_list.md)	 - List .ovhconfig entries
* [ovhcloud webhosting ovh-config recommended](ovhcloud_webhosting_ovh-config_recommended.md)	 - Show recommended values
* [ovhcloud webhosting ovh-config refresh](ovhcloud_webhosting_ovh-config_refresh.md)	 - Refresh cached .ovhconfig data
* [ovhcloud webhosting ovh-config rollback](ovhcloud_webhosting_ovh-config_rollback.md)	 - Rollback a .ovhconfig entry

