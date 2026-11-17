## ovhcloud cloud loadbalancer l7policy edit

Edit a specific L7 policy

### Synopsis

Edit a specific L7 policy, including the rules nested in it.
The flags leave the rules of the policy untouched. To add, change or remove rules, use --editor.

```
ovhcloud cloud loadbalancer l7policy edit <loadbalancer_id> <listener_id> <l7policy_id> [flags]
```

### Options

```
      --action string             L7 policy action (REDIRECT_TO_POOL, REDIRECT_TO_URL, REDIRECT_PREFIX, REJECT)
      --description string        Description
      --editor                    Use a text editor to define parameters
  -h, --help                      help for edit
      --name string               Name of the L7 policy
      --position int              Position on the listener
      --redirect-http-code int    HTTP status code of the redirection (301, 302, 303, 307, 308)
      --redirect-pool-id string   Pool ID to redirect to
      --redirect-prefix string    Prefix URL to redirect to
      --redirect-url string       URL to redirect to
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud loadbalancer l7policy](ovhcloud_cloud_loadbalancer_l7policy.md)	 - Manage L7 policies of loadbalancers

