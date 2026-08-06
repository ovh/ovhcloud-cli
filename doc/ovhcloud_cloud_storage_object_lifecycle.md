## ovhcloud cloud storage object lifecycle

Manage S3™* compatible storage container lifecycle configuration (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

### Options

```
  -h, --help   help for lifecycle
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
  -y, --yes                    Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud storage object](ovhcloud_cloud_storage_object.md)	 - Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage object lifecycle delete](ovhcloud_cloud_storage_object_lifecycle_delete.md)	 - Delete the lifecycle configuration of the given storage container
* [ovhcloud cloud storage object lifecycle edit](ovhcloud_cloud_storage_object_lifecycle_edit.md)	 - Edit the lifecycle configuration of the given storage container
* [ovhcloud cloud storage object lifecycle get](ovhcloud_cloud_storage_object_lifecycle_get.md)	 - Get the lifecycle configuration of the given storage container

