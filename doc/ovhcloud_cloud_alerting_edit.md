## ovhcloud cloud alerting edit

Edit a billing alert configuration

```
ovhcloud cloud alerting edit <alert_id> [flags]
```

### Options

```
      --delay int               Delay between alerts in seconds (minimum 3600)
      --editor                  Use a text editor to define parameters
      --emails strings          Email addresses to receive alerts (comma-separated)
  -h, --help                    help for edit
      --monthly-threshold int   Monthly threshold value
      --name string             Alert name
      --service string          Service of the alert. Allowed: ai_endpoint, all, block_storage, data_platform, instances, instances_gpu, instances_without_gpu, objet_storage, rancher, snapshot
      --status string           Status of the alert. Allowed: deleted, disabled, ok
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

* [ovhcloud cloud alerting](ovhcloud_cloud_alerting.md)	 - Manage billing alert configurations in the given cloud project

