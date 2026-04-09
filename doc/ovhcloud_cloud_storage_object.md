## ovhcloud cloud storage object

Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for object
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud storage](ovhcloud_cloud_storage.md)	 - Manage storage services in the given cloud project
* [ovhcloud cloud storage object add-user](ovhcloud_cloud_storage_object_add-user.md)	 - Add a user to the given storage container with the specified role (admin, deny, readOnly, readWrite)
* [ovhcloud cloud storage object bulk-delete](ovhcloud_cloud_storage_object_bulk-delete.md)	 - Bulk delete objects in the given storage container
* [ovhcloud cloud storage object create](ovhcloud_cloud_storage_object_create.md)	 - Create a new S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object credentials](ovhcloud_cloud_storage_object_credentials.md)	 - Manage storage containers credentials
* [ovhcloud cloud storage object delete](ovhcloud_cloud_storage_object_delete.md)	 - Delete the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object edit](ovhcloud_cloud_storage_object_edit.md)	 - Edit the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object generate-presigned-url](ovhcloud_cloud_storage_object_generate-presigned-url.md)	 - Generate a presigned URL to upload or download an object in the given storage container
* [ovhcloud cloud storage object get](ovhcloud_cloud_storage_object_get.md)	 - Get a specific S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object lifecycle](ovhcloud_cloud_storage_object_lifecycle.md)	 - Manage S3™* compatible storage container lifecycle configuration (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object list](ovhcloud_cloud_storage_object_list.md)	 - List S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object object](ovhcloud_cloud_storage_object_object.md)	 - Manage objects in the given storage container
* [ovhcloud cloud storage object quota](ovhcloud_cloud_storage_object_quota.md)	 - Manage S3™* compatible storage quota (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object replication-job](ovhcloud_cloud_storage_object_replication-job.md)	 - Manage replication jobs for S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

