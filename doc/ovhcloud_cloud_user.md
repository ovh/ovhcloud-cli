## ovhcloud cloud user

Manage users in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for user
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud user create](ovhcloud_cloud_user_create.md)	 - Create a new user
* [ovhcloud cloud user delete](ovhcloud_cloud_user_delete.md)	 - Delete the given user
* [ovhcloud cloud user get](ovhcloud_cloud_user_get.md)	 - Get information about a user
* [ovhcloud cloud user list](ovhcloud_cloud_user_list.md)	 - List users
* [ovhcloud cloud user s3-policy](ovhcloud_cloud_user_s3-policy.md)	 - Manage policies for users on S3™* compatible storage containers (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

