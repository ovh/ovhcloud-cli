## ovhcloud cloud storage swift container edit

Edit the given Swift container

```
ovhcloud cloud storage swift container edit <container_id> [flags]
```

### Options

```
      --cors strings   CORS origins to add to the container (use --editor to replace or remove origins)
      --editor         Use a text editor to define parameters
  -h, --help           help for edit
      --type string    Container type (PRIVATE, PUBLIC, STATIC)
      --wait           Wait for the container to be READY before exiting
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

* [ovhcloud cloud storage swift container](ovhcloud_cloud_storage_swift_container.md)	 - Manage Swift storage containers in the given cloud project

