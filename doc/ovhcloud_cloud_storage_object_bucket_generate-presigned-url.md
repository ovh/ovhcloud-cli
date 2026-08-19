## ovhcloud cloud storage object bucket generate-presigned-url

Generate a presigned URL to upload or download an object in the given storage container

```
ovhcloud cloud storage object bucket generate-presigned-url <container_name> [flags]
```

### Options

```
      --editor                 Use a text editor to define parameters
      --expire int             Expiration time in seconds for the presigned URL (default 60)
      --from-file string       File containing parameters
  -h, --help                   help for generate-presigned-url
      --init-file string       Create a file with example parameters
      --method string          HTTP method for the presigned URL (GET, PUT, DELETE) (default "GET")
      --object string          Name of the object to upload or download
      --replace                Replace parameters file if it already exists
      --storage-class string   Storage class for the object (HIGH_PERF, STANDARD, STANDARD_IA)
      --version-id string      Version ID of the object (if applicable)
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

