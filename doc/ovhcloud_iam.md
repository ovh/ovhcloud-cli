## ovhcloud iam

Manage IAM resources, permissions and policies

### Options

```
  -h, --help   help for iam
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud iam permissions-group](ovhcloud_iam_permissions-group.md)	 - Manage IAM permissions groups
* [ovhcloud iam policy](ovhcloud_iam_policy.md)	 - Manage IAM policies
* [ovhcloud iam resource](ovhcloud_iam_resource.md)	 - Manage IAM resources
* [ovhcloud iam resource-group](ovhcloud_iam_resource-group.md)	 - Manage IAM resource groups

