## ovhcloud cloud storage-s3 bulk-delete

Bulk delete objects in the given storage container

```
ovhcloud cloud storage-s3 bulk-delete [flags]
```

### Options

```
  -h, --help              help for bulk-delete
      --objects strings   List of objects to delete (format is '<object_name>' or '<object_name>:<version_id>'
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud storage-s3](ovhcloud_cloud_storage-s3.md)	 - Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

