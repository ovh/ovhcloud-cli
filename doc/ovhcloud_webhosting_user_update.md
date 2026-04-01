## ovhcloud webhosting user update

Update a FTP/SSH user

```
ovhcloud webhosting user update <service_name> <login> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
  -h, --help               help for update
      --home string        Home directory for the FTP/SSH user
      --password string    FTP/SSH password
      --ssh-state string   SSH state (allowed: active, none)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting user](ovhcloud_webhosting_user.md)	 - Manage FTP/SSH users

