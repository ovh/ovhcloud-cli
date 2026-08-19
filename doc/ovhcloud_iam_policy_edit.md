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
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud iam policy](ovhcloud_iam_policy.md)	 - Manage IAM policies

