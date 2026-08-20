## ovhcloud iam application

Manage the applications your API credentials are issued against

### Options

```
  -h, --help   help for application
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

* [ovhcloud iam](ovhcloud_iam.md)	 - Manage IAM resources, permissions and policies
* [ovhcloud iam application delete](ovhcloud_iam_application_delete.md)	 - Delete an application and every credential issued against it
* [ovhcloud iam application get](ovhcloud_iam_application_get.md)	 - Get one API application
* [ovhcloud iam application list](ovhcloud_iam_application_list.md)	 - List your API applications

