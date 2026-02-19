## ovhcloud domain-name

Retrieve information and manage your domain names

### Options

```
  -h, --help   help for domain-name
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud domain-name edit](ovhcloud_domain-name_edit.md)	 - Edit the given domain name service
* [ovhcloud domain-name get](ovhcloud_domain-name_get.md)	 - Retrieve information of a specific domain name
* [ovhcloud domain-name list](ovhcloud_domain-name_list.md)	 - List your domain names

