## ovhcloud cloud storage-s3 object version

Manage versions of objects in the given storage container

### Options

```
  -h, --help   help for version
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

* [ovhcloud cloud storage-s3 object](ovhcloud_cloud_storage-s3_object.md)	 - Manage objects in the given storage container
* [ovhcloud cloud storage-s3 object version copy](ovhcloud_cloud_storage-s3_object_version_copy.md)	 - Copy a specific version of an object to another bucket or key
* [ovhcloud cloud storage-s3 object version delete](ovhcloud_cloud_storage-s3_object_version_delete.md)	 - Delete a specific version of an object from the storage container
* [ovhcloud cloud storage-s3 object version edit](ovhcloud_cloud_storage-s3_object_version_edit.md)	 - Edit the given version of an object in the storage container
* [ovhcloud cloud storage-s3 object version get](ovhcloud_cloud_storage-s3_object_version_get.md)	 - Get a specific version of an object from the given storage container
* [ovhcloud cloud storage-s3 object version list](ovhcloud_cloud_storage-s3_object_version_list.md)	 - List versions of a specific object in the given storage container
* [ovhcloud cloud storage-s3 object version restore](ovhcloud_cloud_storage-s3_object_version_restore.md)	 - Restore a specific version of an object from archival storage

