## ovhcloud cloud instance snapshot

Manage snapshots of the given instance

### Options

```
  -h, --help   help for snapshot
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
```

### SEE ALSO

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project
* [ovhcloud cloud instance snapshot abort](ovhcloud_cloud_instance_snapshot_abort.md)	 - Abort the snapshot creation of the given instance
* [ovhcloud cloud instance snapshot create](ovhcloud_cloud_instance_snapshot_create.md)	 - Create a snapshot of the given instance
* [ovhcloud cloud instance snapshot delete](ovhcloud_cloud_instance_snapshot_delete.md)	 - Delete a specific instance snapshot in the current cloud project
* [ovhcloud cloud instance snapshot get](ovhcloud_cloud_instance_snapshot_get.md)	 - Get a specific instance snapshot in the current cloud project
* [ovhcloud cloud instance snapshot list](ovhcloud_cloud_instance_snapshot_list.md)	 - List all instance snapshots in the current cloud project

