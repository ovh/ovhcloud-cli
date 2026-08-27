## ovhcloud cloud loadbalancer l7policy create

Create an L7 policy in the given region

```
ovhcloud cloud loadbalancer l7policy create <region> [flags]
```

### Options

```
      --action string             L7 policy action (redirectToPool, redirectToUrl, redirectPrefix, reject)
      --description string        Description
      --editor                    Use a text editor to define parameters
      --from-file string          File containing parameters
  -h, --help                      help for create
      --init-file string          Create a file with example parameters
      --listener-id string        Listener ID
      --name string               Name of the L7 policy
      --position int              Position on the listener
      --redirect-pool-id string   Redirect pool ID
      --redirect-prefix string    Redirect prefix URL
      --redirect-url string       Redirect URL
      --replace                   Replace parameters file if it already exists
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
                               
                               When extracting a single scalar field, the value is printed without surrounding
                               quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud loadbalancer l7policy](ovhcloud_cloud_loadbalancer_l7policy.md)	 - Manage L7 policies of loadbalancers

