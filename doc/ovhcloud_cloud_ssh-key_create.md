## ovhcloud cloud ssh-key create

Create a new SSH key

```
ovhcloud cloud ssh-key create [flags]
```

### Options

```
      --editor              Use a text editor to define parameters
      --from-file string    File containing parameters
  -h, --help                help for create
      --init-file string    Create a file with example parameters
      --name string         Name for the SSH key to create
      --public-key string   Public key for the SSH key to create
      --replace             Replace parameters file if it already exists
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

* [ovhcloud cloud ssh-key](ovhcloud_cloud_ssh-key.md)	 - Manage SSH keys in the given cloud project

