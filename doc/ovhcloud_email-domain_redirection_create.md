## ovhcloud email-domain redirection create

Create a new email redirection

```
ovhcloud email-domain redirection create <service_name> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from string        Source email address (e.g., alias@domain.com)
      --from-file string   File containing parameters
  -h, --help               help for create
      --init-file string   Create a file with example parameters
      --local-copy         Keep a local copy of the email
      --replace            Replace parameters file if it already exists
      --to string          Destination email address
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

* [ovhcloud email-domain redirection](ovhcloud_email-domain_redirection.md)	 - Manage email redirections for your domain

