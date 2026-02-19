## ovhcloud email-domain

Retrieve information and manage your Email Domain services

### Options

```
  -h, --help   help for email-domain
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
* [ovhcloud email-domain get](ovhcloud_email-domain_get.md)	 - Retrieve information of a specific Email Domain
* [ovhcloud email-domain list](ovhcloud_email-domain_list.md)	 - List your Email Domain services
* [ovhcloud email-domain redirection](ovhcloud_email-domain_redirection.md)	 - Manage email redirections for your domain

