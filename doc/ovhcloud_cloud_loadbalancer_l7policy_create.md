## ovhcloud cloud loadbalancer l7policy create

Create an L7 policy on a listener

### Synopsis

Create an L7 policy on a listener.
At most one rule can be given with the --rule-* flags. To create a policy with several rules, use --from-file or --editor.

```
ovhcloud cloud loadbalancer l7policy create <loadbalancer_id> <listener_id> [flags]
```

### Options

```
      --action string              L7 policy action (REDIRECT_TO_POOL, REDIRECT_TO_URL, REDIRECT_PREFIX, REJECT)
      --description string         Description
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Name of the L7 policy
      --position int               Position on the listener
      --redirect-http-code int     HTTP status code of the redirection (301, 302, 303, 307, 308)
      --redirect-pool-id string    Pool ID to redirect to
      --redirect-prefix string     Prefix URL to redirect to
      --redirect-url string        URL to redirect to
      --replace                    Replace parameters file if it already exists
      --rule-compare-type string   Comparison type of the rule (CONTAINS, ENDS_WITH, EQUAL_TO, REGEX, STARTS_WITH)
      --rule-invert                Invert the result of the rule
      --rule-key string            Key of the rule, for COOKIE and HEADER rules
      --rule-type string           Type of the rule (COOKIE, FILE_TYPE, HEADER, HOST_NAME, PATH)
      --rule-value string          Value to compare the rule with
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

