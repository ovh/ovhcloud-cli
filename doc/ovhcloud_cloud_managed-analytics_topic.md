## ovhcloud cloud managed-analytics topic

Manage topics in a specific managed analytics service

### Options

```
  -h, --help   help for topic
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
```

### SEE ALSO

* [ovhcloud cloud managed-analytics](ovhcloud_cloud_managed-analytics.md)	 - Manage managed analytics services in the given cloud project
* [ovhcloud cloud managed-analytics topic create](ovhcloud_cloud_managed-analytics_topic_create.md)	 - Create a new topic in the given managed analytics service
* [ovhcloud cloud managed-analytics topic delete](ovhcloud_cloud_managed-analytics_topic_delete.md)	 - Delete a specific topic in the given managed analytics service
* [ovhcloud cloud managed-analytics topic edit](ovhcloud_cloud_managed-analytics_topic_edit.md)	 - Edit a topic in the given managed analytics service
* [ovhcloud cloud managed-analytics topic get](ovhcloud_cloud_managed-analytics_topic_get.md)	 - Get a specific topic in the given managed analytics service
* [ovhcloud cloud managed-analytics topic list](ovhcloud_cloud_managed-analytics_topic_list.md)	 - List topics in the given managed analytics service

