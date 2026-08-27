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
* [ovhcloud cloud storage object bucket](ovhcloud_cloud_storage_object_bucket.md)	 - Manage object storage buckets in the given cloud project
* [ovhcloud cloud storage object quota](ovhcloud_cloud_storage_object_quota.md)	 - Manage S3™* compatible storage quota (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

