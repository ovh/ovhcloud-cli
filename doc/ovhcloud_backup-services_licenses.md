## ovhcloud backup-services licenses

Show the Veeam licences held by a VSPC tenant

### Options

```
  -h, --help   help for licenses
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
      --tenant string    Backup tenant to work on (default: the only one on the account)
      --vspc string      VSPC tenant to work on (default: the only one in the backup tenant)
```

### SEE ALSO

* [ovhcloud backup-services](ovhcloud_backup-services.md)	 - Retrieve information and manage your Veeam Backup services
* [ovhcloud backup-services licenses list](ovhcloud_backup-services_licenses_list.md)	 - List the Veeam licences
* [ovhcloud backup-services licenses servers](ovhcloud_backup-services_licenses_servers.md)	 - List the backup servers driven by a licence

