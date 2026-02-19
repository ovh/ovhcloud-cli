## ovhcloud iam user

Manage IAM users

### Options

```
  -h, --help   help for user
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

* [ovhcloud iam](ovhcloud_iam.md)	 - Manage IAM resources, permissions and policies
* [ovhcloud iam user create](ovhcloud_iam_user_create.md)	 - Create a new user
* [ovhcloud iam user delete](ovhcloud_iam_user_delete.md)	 - Delete a specific IAM user
* [ovhcloud iam user edit](ovhcloud_iam_user_edit.md)	 - Edit an existing user
* [ovhcloud iam user get](ovhcloud_iam_user_get.md)	 - Get a specific IAM user
* [ovhcloud iam user list](ovhcloud_iam_user_list.md)	 - List IAM users
* [ovhcloud iam user token](ovhcloud_iam_user_token.md)	 - Manage IAM user tokens

