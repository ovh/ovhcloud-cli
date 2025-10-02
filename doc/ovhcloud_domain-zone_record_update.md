## ovhcloud domain-zone record update

Update a single DNS record from your zone

```
ovhcloud domain-zone record update <zone_name> <record_id> [flags]
```

### Options

```
  -h, --help               help for update
      --subdomain string   Subdomain to update
      --target string      New target to apply
      --ttl int            New TTL to apply
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

* [ovhcloud domain-zone record](ovhcloud_domain-zone_record.md)	 - Retrieve information and manage your DNS record within a zone

