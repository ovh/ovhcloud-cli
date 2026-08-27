## ovhcloud cloud storage block volume

Manage block storage volumes in the given cloud project

### Options

```
  -h, --help   help for volume
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

* [ovhcloud cloud storage block](ovhcloud_cloud_storage_block.md)	 - Manage block storage in the given cloud project
* [ovhcloud cloud storage block volume attach](ovhcloud_cloud_storage_block_volume_attach.md)	 - Attach the given volume to the given instance
* [ovhcloud cloud storage block volume create](ovhcloud_cloud_storage_block_volume_create.md)	 - Create a new volume
* [ovhcloud cloud storage block volume delete](ovhcloud_cloud_storage_block_volume_delete.md)	 - Delete the given volume
* [ovhcloud cloud storage block volume detach](ovhcloud_cloud_storage_block_volume_detach.md)	 - Detach the given volume from the given instance
* [ovhcloud cloud storage block volume edit](ovhcloud_cloud_storage_block_volume_edit.md)	 - Edit the given volume
* [ovhcloud cloud storage block volume get](ovhcloud_cloud_storage_block_volume_get.md)	 - Get a specific volume
* [ovhcloud cloud storage block volume list](ovhcloud_cloud_storage_block_volume_list.md)	 - List volumes

