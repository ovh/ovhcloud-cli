## ovhcloud webhosting user create

Create a FTP/SSH user

```
ovhcloud webhosting user create <service_name> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for create
      --home string        Home directory for the FTP/SSH user
      --login string       FTP/SSH login
      --password string    FTP/SSH password
      --ssh-state string   SSH state (allowed: active, none, sftponly)
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
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting user](ovhcloud_webhosting_user.md)	 - Manage FTP/SSH users

