## ovhcloud iam policy edit

Edit specific IAM policy

```
ovhcloud iam policy edit <policy_id> [flags]
```

### Options

```
      --allow strings               List of allowed actions
      --deny strings                List of denied actions
      --description string          Description of the policy
      --editor                      Use a text editor to define parameters
      --except strings              List of actions to filter from the allowed list
      --expiredAt string            Expiration date of the policy (RFC3339 format), after this date it will no longer be applied
  -h, --help                        help for edit
      --identity strings            Identities to which the policy applies
      --name string                 Name of the policy
      --permissions-group strings   Permissions group URNs
      --resource strings            Resource URNs
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

* [ovhcloud iam policy](ovhcloud_iam_policy.md)	 - Manage IAM policies

