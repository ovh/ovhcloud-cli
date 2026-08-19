## ovhcloud ip mitigation-profile

Manage how long auto-mitigation stays on after an attack

### Options

```
  -h, --help   help for mitigation-profile
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
* [ovhcloud ip mitigation-profile delete](ovhcloud_ip_mitigation-profile_delete.md)	 - Delete an auto-mitigation profile
* [ovhcloud ip mitigation-profile get](ovhcloud_ip_mitigation-profile_get.md)	 - Show one auto-mitigation profile
* [ovhcloud ip mitigation-profile list](ovhcloud_ip_mitigation-profile_list.md)	 - List the auto-mitigation profiles of this block
* [ovhcloud ip mitigation-profile set](ovhcloud_ip_mitigation-profile_set.md)	 - Set the auto-mitigation delay of an address, creating its profile if needed

