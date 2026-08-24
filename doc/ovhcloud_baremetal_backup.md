## ovhcloud baremetal backup

Manage the backup spaces of a dedicated server

### Options

```
  -h, --help   help for backup
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

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services
* [ovhcloud baremetal backup cloud](ovhcloud_baremetal_backup_cloud.md)	 - Manage the cloud backup containers of the server
* [ovhcloud baremetal backup ftp](ovhcloud_baremetal_backup_ftp.md)	 - Manage the Backup FTP space included with the server
* [ovhcloud baremetal backup orderable](ovhcloud_baremetal_backup_orderable.md)	 - Show the backup storage capacities this server accepts

