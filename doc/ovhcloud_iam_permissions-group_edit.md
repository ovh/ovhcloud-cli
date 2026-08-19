## ovhcloud iam permissions-group edit

Edit a specific IAM permissions group

```
ovhcloud iam permissions-group edit <permissions_group_id> [flags]
```

### Options

```
      --allow strings        List of allowed actions
      --deny strings         List of denied actions
      --description string   Description of the policy
      --editor               Use a text editor to define parameters
      --except strings       List of actions to filter from the allowed list
  -h, --help                 help for edit
      --name string          Name of the policy
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud iam permissions-group](ovhcloud_iam_permissions-group.md)	 - Manage IAM permissions groups

