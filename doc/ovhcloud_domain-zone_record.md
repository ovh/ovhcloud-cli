## ovhcloud domain-zone record

Retrieve information and manage your DNS records within a zone

### Options

```
  -h, --help   help for record
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

* [ovhcloud domain-zone](ovhcloud_domain-zone.md)	 - Retrieve information and manage your domain zones
* [ovhcloud domain-zone record create](ovhcloud_domain-zone_record_create.md)	 - Create a single DNS record in your zone
* [ovhcloud domain-zone record delete](ovhcloud_domain-zone_record_delete.md)	 - Delete a single DNS record from your zone
* [ovhcloud domain-zone record get](ovhcloud_domain-zone_record_get.md)	 - Get a single DNS record from your zone
* [ovhcloud domain-zone record list](ovhcloud_domain-zone_record_list.md)	 - List all DNS records from your zone
* [ovhcloud domain-zone record update](ovhcloud_domain-zone_record_update.md)	 - Update a single DNS record from your zone

