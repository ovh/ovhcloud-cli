## ovhcloud cloud storage-block backup

Manage volume backups in the given cloud project

### Options

```
  -h, --help   help for backup
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud storage-block](ovhcloud_cloud_storage-block.md)	 - Manage block storage volumes in the given cloud project
* [ovhcloud cloud storage-block backup create](ovhcloud_cloud_storage-block_backup_create.md)	 - Create a backup of the given volume
* [ovhcloud cloud storage-block backup delete](ovhcloud_cloud_storage-block_backup_delete.md)	 - Delete the given volume backup
* [ovhcloud cloud storage-block backup get](ovhcloud_cloud_storage-block_backup_get.md)	 - Get a specific volume backup
* [ovhcloud cloud storage-block backup list](ovhcloud_cloud_storage-block_backup_list.md)	 - List volume backups
* [ovhcloud cloud storage-block backup restore](ovhcloud_cloud_storage-block_backup_restore.md)	 - Restore a volume from the given backup

