## ovhcloud cloud managed-analytics user edit

Edit a user in the given managed analytics service

### Synopsis

Use this command to edit a user in the given managed analytics service.
There are two ways to define the edition parameters:

1. Using only CLI flags:

  For Clickhouse engine:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --roles role1,role2

  For OpenSearch engine:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --acls "logs-.*:write" --acls "metrics-.*:read"

2. Using your default text editor:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --editor
  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-analytics user edit <service_id> <user_id> --editor --roles role1,role2


```
ovhcloud cloud managed-analytics user edit <service_id> <user_id> [flags]
```

### Options

```
      --acls stringArray   ACL granted to the user (opensearch only, format: pattern:permission)
      --editor             Use a text editor to define parameters
  -h, --help               help for edit
      --roles strings      Roles granted to the user (clickhouse only)
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

* [ovhcloud cloud managed-analytics user](ovhcloud_cloud_managed-analytics_user.md)	 - Manage users in a specific managed analytics service

