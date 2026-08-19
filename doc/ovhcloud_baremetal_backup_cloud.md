## ovhcloud baremetal backup cloud

Manage the cloud backup containers of the server

### Options

```
  -h, --help   help for cloud
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
* [ovhcloud baremetal backup cloud create](ovhcloud_baremetal_backup_cloud_create.md)	 - Create the cloud backup containers of this server
* [ovhcloud baremetal backup cloud delete](ovhcloud_baremetal_backup_cloud_delete.md)	 - Deactivate the cloud backup — the container data is kept
* [ovhcloud baremetal backup cloud offer](ovhcloud_baremetal_backup_cloud_offer.md)	 - Show what a cloud backup would hold for this server
* [ovhcloud baremetal backup cloud password](ovhcloud_baremetal_backup_cloud_password.md)	 - Reset the four cloud backup passwords
* [ovhcloud baremetal backup cloud show](ovhcloud_baremetal_backup_cloud_show.md)	 - Show the cloud backup containers of this server

