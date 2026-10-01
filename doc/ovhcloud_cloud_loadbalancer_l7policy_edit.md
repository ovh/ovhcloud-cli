## ovhcloud cloud loadbalancer l7policy edit

Edit a specific L7 policy

```
ovhcloud cloud loadbalancer l7policy edit <l7policy_id> [flags]
```

### Options

```
      --action string             L7 policy action
      --description string        Description
      --editor                    Use a text editor to define parameters
  -h, --help                      help for edit
      --listener-id string        Listener ID
      --name string               Name of the L7 policy
      --position int              Position on the listener
      --redirect-pool-id string   Redirect pool ID
      --redirect-prefix string    Redirect prefix URL
      --redirect-url string       Redirect URL
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

* [ovhcloud cloud loadbalancer l7policy](ovhcloud_cloud_loadbalancer_l7policy.md)	 - Manage L7 policies of loadbalancers

