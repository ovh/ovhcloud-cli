## ovhcloud cloud instance shelve

Shelve the given instance

### Synopsis

The resources dedicated to the Public Cloud instance are released.
The data of the local storage will be stored, the duration of the operation depends on the size of the local disk.
The instance can be unshelved at any time. Meanwhile hourly instances will not be billed.
The Snapshot Storage used to store the instance's data will be billed.

```
ovhcloud cloud instance shelve <instance_id> [flags]
```

### Options

```
  -h, --help   help for shelve
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

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project

