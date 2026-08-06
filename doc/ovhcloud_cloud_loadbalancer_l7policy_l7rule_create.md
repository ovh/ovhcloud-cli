## ovhcloud cloud loadbalancer l7policy l7rule create

Create an L7 rule in a specific L7 policy

```
ovhcloud cloud loadbalancer l7policy l7rule create <l7policy_id> [flags]
```

### Options

```
      --compare-type string   Comparison type (contains, endsWith, equalTo, regex, startsWith)
      --editor                Use a text editor to define parameters
      --from-file string      File containing parameters
  -h, --help                  help for create
      --init-file string      Create a file with example parameters
      --key string            Key to use for comparison
      --replace               Replace parameters file if it already exists
      --rule-type string      Rule type (cookie, fileType, header, hostName, path, sslConnHasCert, sslDNField, sslVerifyResult)
      --value string          Value to compare
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
  -y, --yes                    Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud loadbalancer l7policy l7rule](ovhcloud_cloud_loadbalancer_l7policy_l7rule.md)	 - Manage L7 rules of a loadbalancer L7 policy

