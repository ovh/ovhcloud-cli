## ovhcloud iam permissions-group

Manage IAM permissions groups

### Options

```
  -h, --help   help for permissions-group
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

* [ovhcloud iam](ovhcloud_iam.md)	 - Manage IAM resources, permissions and policies
* [ovhcloud iam permissions-group edit](ovhcloud_iam_permissions-group_edit.md)	 - Edit a specific IAM permissions group
* [ovhcloud iam permissions-group get](ovhcloud_iam_permissions-group_get.md)	 - Get a specific IAM permissions group
* [ovhcloud iam permissions-group list](ovhcloud_iam_permissions-group_list.md)	 - List IAM permissions groups

