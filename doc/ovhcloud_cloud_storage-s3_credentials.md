## ovhcloud cloud storage-s3 credentials

Manage storage containers credentials

### Options

```
  -h, --help   help for credentials
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

* [ovhcloud cloud storage-s3](ovhcloud_cloud_storage-s3.md)	 - Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage-s3 credentials create](ovhcloud_cloud_storage-s3_credentials_create.md)	 - Create credentials for the given user ID
* [ovhcloud cloud storage-s3 credentials delete](ovhcloud_cloud_storage-s3_credentials_delete.md)	 - Delete credentials for the given user ID and access ID
* [ovhcloud cloud storage-s3 credentials get](ovhcloud_cloud_storage-s3_credentials_get.md)	 - Get credentials for the given user ID and access ID
* [ovhcloud cloud storage-s3 credentials list](ovhcloud_cloud_storage-s3_credentials_list.md)	 - List credentials for the given user ID

