## ovhcloud cloud storage-s3 edit

Edit the given S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

```
ovhcloud cloud storage-s3 edit <container_name> [flags]
```

### Options

```
      --editor                            Use a text editor to define parameters
      --encryption-sse-algorithm string   Encryption SSE Algorithm (AES256, plaintext)
  -h, --help                              help for edit
      --object-lock-rule-mode string      Object lock mode (compliance, governance)
      --object-lock-rule-period string    Object lock period (e.g., P3Y6M4DT12H30M5S)
      --object-lock-status string         Object lock status (disabled, enabled)
      --tag stringToString                Container tags as key=value pairs (default [])
      --versioning-status string          Versioning status (disabled, enabled, suspended)
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

