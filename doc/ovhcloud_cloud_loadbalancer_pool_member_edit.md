## ovhcloud cloud loadbalancer pool member edit

Edit a specific pool member

```
ovhcloud cloud loadbalancer pool member edit <pool_id> <member_id> [flags]
```

### Options

```
      --editor        Use a text editor to define parameters
  -h, --help          help for edit
      --name string   Name of the member
      --weight int    Weight of the member (1-256)
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

* [ovhcloud cloud loadbalancer pool member](ovhcloud_cloud_loadbalancer_pool_member.md)	 - Manage members of a loadbalancer pool

