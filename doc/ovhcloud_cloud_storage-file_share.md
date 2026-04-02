## ovhcloud cloud storage-file share

Manage file storage shares

### Options

```
  -h, --help   help for share
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
      --region string          Region (skip region discovery if set)
```

### SEE ALSO

* [ovhcloud cloud storage-file](ovhcloud_cloud_storage-file.md)	 - Manage file storage shares in the given cloud project
* [ovhcloud cloud storage-file share acl](ovhcloud_cloud_storage-file_share_acl.md)	 - Manage share access control lists
* [ovhcloud cloud storage-file share create](ovhcloud_cloud_storage-file_share_create.md)	 - Create a new share
* [ovhcloud cloud storage-file share delete](ovhcloud_cloud_storage-file_share_delete.md)	 - Delete the given share
* [ovhcloud cloud storage-file share edit](ovhcloud_cloud_storage-file_share_edit.md)	 - Edit the given share
* [ovhcloud cloud storage-file share get](ovhcloud_cloud_storage-file_share_get.md)	 - Get a specific share
* [ovhcloud cloud storage-file share list](ovhcloud_cloud_storage-file_share_list.md)	 - List shares
* [ovhcloud cloud storage-file share snapshot](ovhcloud_cloud_storage-file_share_snapshot.md)	 - Manage share snapshots

