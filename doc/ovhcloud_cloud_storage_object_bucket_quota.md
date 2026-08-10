## ovhcloud cloud storage object bucket quota

Manage S3™* compatible storage quota (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

### Options

```
  -h, --help   help for quota
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
* [ovhcloud cloud storage object bucket quota delete](ovhcloud_cloud_storage_object_bucket_quota_delete.md)	 - Delete storage quota for the given region
* [ovhcloud cloud storage object bucket quota edit](ovhcloud_cloud_storage_object_bucket_quota_edit.md)	 - Edit storage quota for the given region
* [ovhcloud cloud storage object bucket quota get](ovhcloud_cloud_storage_object_bucket_quota_get.md)	 - Get storage quota for the given region

