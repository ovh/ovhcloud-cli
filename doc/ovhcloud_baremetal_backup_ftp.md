## ovhcloud baremetal backup ftp

Manage the Backup FTP space included with the server

### Options

```
  -h, --help   help for ftp
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

* [ovhcloud baremetal backup](ovhcloud_baremetal_backup.md)	 - Manage the backup spaces of a dedicated server
* [ovhcloud baremetal backup ftp acl](ovhcloud_baremetal_backup_ftp_acl.md)	 - Manage who may reach the Backup FTP space, and how
* [ovhcloud baremetal backup ftp authorizable-blocks](ovhcloud_baremetal_backup_ftp_authorizable-blocks.md)	 - List the IP blocks that may be allowed on this Backup FTP space
* [ovhcloud baremetal backup ftp create](ovhcloud_baremetal_backup_ftp_create.md)	 - Create the Backup FTP space included with this server
* [ovhcloud baremetal backup ftp delete](ovhcloud_baremetal_backup_ftp_delete.md)	 - Terminate the Backup FTP space — all data is permanently deleted
* [ovhcloud baremetal backup ftp password](ovhcloud_baremetal_backup_ftp_password.md)	 - Change the Backup FTP password (the new one is emailed)
* [ovhcloud baremetal backup ftp show](ovhcloud_baremetal_backup_ftp_show.md)	 - Show the Backup FTP space of this server

