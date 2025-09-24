## ovhcloud cloud storage-block snapshot

Manage snapshots of the given volume

### Options

```
  -h, --help   help for snapshot
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
* [ovhcloud cloud storage-block snapshot create](ovhcloud_cloud_storage-block_snapshot_create.md)	 - Create a snapshot of the given volume
* [ovhcloud cloud storage-block snapshot delete](ovhcloud_cloud_storage-block_snapshot_delete.md)	 - Delete the given snapshot
* [ovhcloud cloud storage-block snapshot list](ovhcloud_cloud_storage-block_snapshot_list.md)	 - List snapshots of the given volume

