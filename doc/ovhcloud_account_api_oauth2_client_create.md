## ovhcloud account api oauth2 client create

Create a new OAuth2 client

```
ovhcloud account api oauth2 client create [flags]
```

### Options

```
      --callback-urls stringArray   Callback URLs for the OAuth2 client
      --description string          Description of the OAuth2 client
      --editor                      Use a text editor to define parameters
      --flow string                 OAuth2 flow type (default: AUTHORIZATION_CODE) (default "AUTHORIZATION_CODE")
      --from-file string            File containing parameters
  -h, --help                        help for create
      --init-file string            Create a file with example parameters
      --name string                 Name of the OAuth2 client
      --replace                     Replace parameters file if it already exists
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

* [ovhcloud account api oauth2 client](ovhcloud_account_api_oauth2_client.md)	 - Manage your OAuth2 clients

