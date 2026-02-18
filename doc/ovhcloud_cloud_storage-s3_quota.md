## ovhcloud cloud storage-s3 quota

Manage S3™* compatible storage quota (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

### Options

```
  -h, --help   help for quota
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

* [ovhcloud cloud storage-s3](ovhcloud_cloud_storage-s3.md)	 - Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage-s3 quota delete](ovhcloud_cloud_storage-s3_quota_delete.md)	 - Delete storage quota for the given region
* [ovhcloud cloud storage-s3 quota edit](ovhcloud_cloud_storage-s3_quota_edit.md)	 - Edit storage quota for the given region
* [ovhcloud cloud storage-s3 quota get](ovhcloud_cloud_storage-s3_quota_get.md)	 - Get storage quota for the given region

