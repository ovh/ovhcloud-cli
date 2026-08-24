## ovhcloud baremetal backup ftp acl

Manage who may reach the Backup FTP space, and how

### Options

```
  -h, --help   help for acl
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

* [ovhcloud baremetal backup ftp](ovhcloud_baremetal_backup_ftp.md)	 - Manage the Backup FTP space included with the server
* [ovhcloud baremetal backup ftp acl add](ovhcloud_baremetal_backup_ftp_acl_add.md)	 - Allow an IP block to reach the Backup FTP space
* [ovhcloud baremetal backup ftp acl delete](ovhcloud_baremetal_backup_ftp_acl_delete.md)	 - Stop an IP block reaching the Backup FTP space
* [ovhcloud baremetal backup ftp acl get](ovhcloud_baremetal_backup_ftp_acl_get.md)	 - Show one access rule
* [ovhcloud baremetal backup ftp acl list](ovhcloud_baremetal_backup_ftp_acl_list.md)	 - List the access rules of the Backup FTP space
* [ovhcloud baremetal backup ftp acl set](ovhcloud_baremetal_backup_ftp_acl_set.md)	 - Change the protocols an access rule opens

