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
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-analytics topic](ovhcloud_cloud_managed-analytics_topic.md)	 - Manage topics in a specific managed analytics service

