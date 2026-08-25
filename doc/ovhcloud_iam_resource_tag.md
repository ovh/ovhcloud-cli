## ovhcloud iam resource tag

Add or remove resource tags without touching the others

### Options

```
  -h, --help   help for tag
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud iam resource](ovhcloud_iam_resource.md)	 - Manage IAM resources
* [ovhcloud iam resource tag remove](ovhcloud_iam_resource_tag_remove.md)	 - Remove tags by key, leaving the other tags in place
* [ovhcloud iam resource tag set](ovhcloud_iam_resource_tag_set.md)	 - Add or update tags, leaving the other tags in place

