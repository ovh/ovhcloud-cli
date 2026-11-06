## ovhcloud cloud storage object bucket add-user

Add a user to the given storage container with the specified role (admin, deny, readOnly, readWrite)

```
ovhcloud cloud storage object bucket add-user <container_name> <user_id> <role (admin, deny, readOnly, readWrite)> [flags]
```

### Options

```
  -h, --help   help for add-user
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

