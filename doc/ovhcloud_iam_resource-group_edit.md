## ovhcloud iam resource-group edit

Edit a specific IAM resource group

```
ovhcloud iam resource-group edit <resource_group_id> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
  -h, --help               help for edit
      --name string        Name of the resource group
      --resource strings   List of resource URNs to include in the group
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

* [ovhcloud iam resource-group](ovhcloud_iam_resource-group.md)	 - Manage IAM resource groups

