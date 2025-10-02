## ovhcloud domain-zone record

Retrieve information and manage your DNS record within a zone

### Options

```
  -h, --help   help for record
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

* [ovhcloud domain-zone](ovhcloud_domain-zone.md)	 - Retrieve information and manage your domain zones
* [ovhcloud domain-zone record get](ovhcloud_domain-zone_record_get.md)	 - Get a single DNS record from your zone
* [ovhcloud domain-zone record update](ovhcloud_domain-zone_record_update.md)	 - Update a single DNS record from your zone

