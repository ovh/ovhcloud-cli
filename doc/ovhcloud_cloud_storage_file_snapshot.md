## ovhcloud cloud storage file snapshot

Manage file storage snapshots

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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud storage file](ovhcloud_cloud_storage_file.md)	 - Manage file storage in the given cloud project
* [ovhcloud cloud storage file snapshot create](ovhcloud_cloud_storage_file_snapshot_create.md)	 - Create a new file storage snapshot
* [ovhcloud cloud storage file snapshot delete](ovhcloud_cloud_storage_file_snapshot_delete.md)	 - Delete the given file storage snapshot
* [ovhcloud cloud storage file snapshot edit](ovhcloud_cloud_storage_file_snapshot_edit.md)	 - Edit the given file storage snapshot
* [ovhcloud cloud storage file snapshot get](ovhcloud_cloud_storage_file_snapshot_get.md)	 - Get a specific file storage snapshot
* [ovhcloud cloud storage file snapshot list](ovhcloud_cloud_storage_file_snapshot_list.md)	 - List file storage snapshots

