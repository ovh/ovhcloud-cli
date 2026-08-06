## ovhcloud cloud managed-analytics topic create

Create a new topic in the given managed analytics service

```
ovhcloud cloud managed-analytics topic create <service_id> [flags]
```

### Options

```
      --editor                    Use a text editor to define parameters
  -h, --help                      help for create
      --min-insync-replicas int   Minimum in-sync replicas
      --name string               Topic name
      --partitions int            Number of partitions
      --replication int           Number of replications
      --retention-bytes int       Retention size in bytes (-1 for unlimited)
      --retention-hours int       Retention duration in hours (-1 for unlimited)
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

* [ovhcloud cloud managed-analytics topic](ovhcloud_cloud_managed-analytics_topic.md)	 - Manage topics in a specific managed analytics service

