## ovhcloud domain-zone

Retrieve information and manage your domain zones

### Options

```
  -h, --help   help for domain-zone
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
* [ovhcloud domain-zone get](ovhcloud_domain-zone_get.md)	 - Retrieve information of a specific domain zone
* [ovhcloud domain-zone list](ovhcloud_domain-zone_list.md)	 - List your domain zones
* [ovhcloud domain-zone record](ovhcloud_domain-zone_record.md)	 - Retrieve information and manage your DNS records within a zone
* [ovhcloud domain-zone refresh](ovhcloud_domain-zone_refresh.md)	 - Refresh the given zone

