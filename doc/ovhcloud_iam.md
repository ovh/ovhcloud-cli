## ovhcloud iam

Manage IAM resources, permissions and policies

### Options

```
  -h, --help   help for iam
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud iam permissions-group](ovhcloud_iam_permissions-group.md)	 - Manage IAM permissions groups
* [ovhcloud iam policy](ovhcloud_iam_policy.md)	 - Manage IAM policies
* [ovhcloud iam resource](ovhcloud_iam_resource.md)	 - Manage IAM resources
* [ovhcloud iam resource-group](ovhcloud_iam_resource-group.md)	 - Manage IAM resource groups
* [ovhcloud iam user](ovhcloud_iam_user.md)	 - Manage IAM users

