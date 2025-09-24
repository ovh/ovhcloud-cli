## ovhcloud vps automated-backup

Manage VPS automated backups

### Options

```
  -h, --help   help for automated-backup
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

* [ovhcloud vps](ovhcloud_vps.md)	 - Retrieve information and manage your VPS services
* [ovhcloud vps automated-backup get-config](ovhcloud_vps_automated-backup_get-config.md)	 - Retrieve automated backup configuration of the given VPS
* [ovhcloud vps automated-backup list](ovhcloud_vps_automated-backup_list.md)	 - List all automated backups of the given VPS
* [ovhcloud vps automated-backup list-restore-points](ovhcloud_vps_automated-backup_list-restore-points.md)	 - List all restore points of the given VPS
* [ovhcloud vps automated-backup reschedule](ovhcloud_vps_automated-backup_reschedule.md)	 - Reschedule the automated backup of the given VPS
* [ovhcloud vps automated-backup restore](ovhcloud_vps_automated-backup_restore.md)	 - Restore the automated backup of the given VPS

