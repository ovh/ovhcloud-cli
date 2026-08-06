## ovhcloud cloud loadbalancer l7policy l7rule list

List L7 rules of a specific L7 policy

```
ovhcloud cloud loadbalancer l7policy l7rule list <l7policy_id> [flags]
```

### Options

```
      --filter stringArray   Filter results by any property using https://github.com/PaesslerAG/gval syntax
                             Examples:
                               --filter 'state=="running"'
                               --filter 'name=~"^my.*"'
                               --filter 'nested.property.subproperty>10'
                               --filter 'startDate>="2023-12-01"'
                               --filter 'name=~"something" && nbField>10'
  -h, --help                 help for list
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

