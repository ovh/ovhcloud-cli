## ovhcloud cloud storage object bucket edit

Edit the given S3™* compatible bucket (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

```
ovhcloud cloud storage object bucket edit <bucket_id> [flags]
```

### Options

```
      --editor                           Use a text editor to define parameters
      --encryption-algorithm string      Server-side encryption algorithm (AES256, PLAINTEXT)
  -h, --help                             help for edit
      --object-lock-mode string          Object lock retention mode (COMPLIANCE, GOVERNANCE)
      --object-lock-retention-days int   Number of days to retain objects
      --owner-user-id string             Owner user ID of the bucket
      --tag stringToString               Bucket tags as key=value pairs (default [])
      --versioning-status string         Versioning status (DISABLED, ENABLED, SUSPENDED)
      --wait                             Wait for the bucket to be ready before exiting
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud storage object bucket](ovhcloud_cloud_storage_object_bucket.md)	 - Manage object storage buckets in the given cloud project

