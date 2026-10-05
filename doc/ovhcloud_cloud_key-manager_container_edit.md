## ovhcloud cloud key-manager container edit

Edit the given Key Manager container (only secret references are mutable)

```
ovhcloud cloud key-manager container edit <container_id> [flags]
```

### Options

```
      --editor                   Use a text editor to define parameters
  -h, --help                     help for edit
      --secret-ref stringArray   Secret reference as '<name>=<secretId>' (repeatable, replaces all existing references)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud key-manager container](ovhcloud_cloud_key-manager_container.md)	 - Manage Key Manager containers

