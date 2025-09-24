## ovhcloud vps snapshot

Manage VPS snapshots

### Options

```
  -h, --help   help for snapshot
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
* [ovhcloud vps snapshot abort](ovhcloud_vps_snapshot_abort.md)	 - Abort the creation of a VPS snapshot
* [ovhcloud vps snapshot create](ovhcloud_vps_snapshot_create.md)	 - Create a snapshot of the given VPS
* [ovhcloud vps snapshot delete](ovhcloud_vps_snapshot_delete.md)	 - Delete the given VPS snapshot
* [ovhcloud vps snapshot download](ovhcloud_vps_snapshot_download.md)	 - Download the snapshot of the given VPS
* [ovhcloud vps snapshot edit](ovhcloud_vps_snapshot_edit.md)	 - Edit the given VPS snapshot
* [ovhcloud vps snapshot get](ovhcloud_vps_snapshot_get.md)	 - Retrieve information of a specific VPS snapshot
* [ovhcloud vps snapshot restore](ovhcloud_vps_snapshot_restore.md)	 - Restore the snapshot of the given VPS

