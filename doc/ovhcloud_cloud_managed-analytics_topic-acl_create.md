## ovhcloud cloud managed-analytics topic-acl create

Create a new topic ACL in the given managed analytics service

### Synopsis

Use this command to create a topic ACL in the given managed analytics service (kafka only).

	ovhcloud cloud managed-analytics topic-acl create <service_id> --permission read --topic my-topic --username myuser


```
ovhcloud cloud managed-analytics topic-acl create <service_id> [flags]
```

### Options

```
  -h, --help                help for create
      --permission string   ACL permission (e.g. read, write, readwrite)
      --topic string        Topic name the ACL applies to
      --username string     Username the ACL applies to
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
                               
                               When extracting a single scalar field, the value is printed without surrounding
                               quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-analytics topic-acl](ovhcloud_cloud_managed-analytics_topic-acl.md)	 - Manage topic ACLs in a specific managed analytics service

