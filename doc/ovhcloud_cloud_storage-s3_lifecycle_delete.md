## ovhcloud cloud storage-s3 lifecycle delete

Delete the lifecycle configuration of the given storage container

```
ovhcloud cloud storage-s3 lifecycle delete <container_name> [flags]
```

### Options

```
  -h, --help   help for delete
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

* [ovhcloud cloud storage-s3 lifecycle](ovhcloud_cloud_storage-s3_lifecycle.md)	 - Manage S3™* compatible storage container lifecycle configuration (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

