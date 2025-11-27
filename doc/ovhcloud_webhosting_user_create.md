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
      --ssh-state string   SSH state (allowed: active, none)
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

* [ovhcloud webhosting user](ovhcloud_webhosting_user.md)	 - Create and manage the FTP/SSH users allowed to access your web hosting space.
