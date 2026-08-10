## ovhcloud cloud storage object bucket object

Manage objects in the given storage container

### Options

```
  -h, --help   help for object
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

* [ovhcloud cloud storage object bucket](ovhcloud_cloud_storage_object_bucket.md)	 - Manage object storage buckets in the given cloud project
* [ovhcloud cloud storage object bucket object copy](ovhcloud_cloud_storage_object_bucket_object_copy.md)	 - Copy the given object to another bucket or key
* [ovhcloud cloud storage object bucket object delete](ovhcloud_cloud_storage_object_bucket_object_delete.md)	 - Delete the given object from the storage container
* [ovhcloud cloud storage object bucket object edit](ovhcloud_cloud_storage_object_bucket_object_edit.md)	 - Edit the given object in the storage container
* [ovhcloud cloud storage object bucket object get](ovhcloud_cloud_storage_object_bucket_object_get.md)	 - Get a specific object from the given storage container
* [ovhcloud cloud storage object bucket object list](ovhcloud_cloud_storage_object_bucket_object_list.md)	 - List objects in the given storage container
* [ovhcloud cloud storage object bucket object restore](ovhcloud_cloud_storage_object_bucket_object_restore.md)	 - Restore the given object from archival storage
* [ovhcloud cloud storage object bucket object version](ovhcloud_cloud_storage_object_bucket_object_version.md)	 - Manage versions of objects in the given storage container

