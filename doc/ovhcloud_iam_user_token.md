## ovhcloud iam user token

Manage IAM user tokens

### Options

```
  -h, --help   help for token
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

* [ovhcloud iam user](ovhcloud_iam_user.md)	 - Manage IAM users
* [ovhcloud iam user token create](ovhcloud_iam_user_token_create.md)	 - Create a new token
* [ovhcloud iam user token delete](ovhcloud_iam_user_token_delete.md)	 - Delete a specific token of an IAM user
* [ovhcloud iam user token get](ovhcloud_iam_user_token_get.md)	 - Get a specific token of an IAM user
* [ovhcloud iam user token list](ovhcloud_iam_user_token_list.md)	 - List tokens of a specific IAM user

