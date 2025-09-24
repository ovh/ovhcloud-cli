## ovhcloud cloud storage-block

Manage block storage volumes in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for storage-block
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe
* [ovhcloud cloud storage-block attach](ovhcloud_cloud_storage-block_attach.md)	 - Attach the given volume to the given instance
* [ovhcloud cloud storage-block backup](ovhcloud_cloud_storage-block_backup.md)	 - Manage volume backups in the given cloud project
* [ovhcloud cloud storage-block create](ovhcloud_cloud_storage-block_create.md)	 - Create a new volume
* [ovhcloud cloud storage-block create-from-backup](ovhcloud_cloud_storage-block_create-from-backup.md)	 - Create a volume from the given backup
* [ovhcloud cloud storage-block delete](ovhcloud_cloud_storage-block_delete.md)	 - Delete the given volume
* [ovhcloud cloud storage-block detach](ovhcloud_cloud_storage-block_detach.md)	 - Detach the given volume from the given instance
* [ovhcloud cloud storage-block edit](ovhcloud_cloud_storage-block_edit.md)	 - Edit the given volume
* [ovhcloud cloud storage-block get](ovhcloud_cloud_storage-block_get.md)	 - Get a specific volume
* [ovhcloud cloud storage-block list](ovhcloud_cloud_storage-block_list.md)	 - List volumes
* [ovhcloud cloud storage-block snapshot](ovhcloud_cloud_storage-block_snapshot.md)	 - Manage snapshots of the given volume
* [ovhcloud cloud storage-block upsize](ovhcloud_cloud_storage-block_upsize.md)	 - Upsize the given volume

