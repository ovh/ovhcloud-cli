## ovhcloud email-domain redirection

Manage email redirections for your domain

### Options

```
  -h, --help   help for redirection
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

* [ovhcloud email-domain](ovhcloud_email-domain.md)	 - Retrieve information and manage your Email Domain services
* [ovhcloud email-domain redirection create](ovhcloud_email-domain_redirection_create.md)	 - Create a new email redirection
* [ovhcloud email-domain redirection delete](ovhcloud_email-domain_redirection_delete.md)	 - Delete an email redirection
* [ovhcloud email-domain redirection get](ovhcloud_email-domain_redirection_get.md)	 - Get details of a specific email redirection
* [ovhcloud email-domain redirection list](ovhcloud_email-domain_redirection_list.md)	 - List all email redirections for a domain

