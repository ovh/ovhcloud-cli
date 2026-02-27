## ovhcloud cloud loadbalancer l7policy l7rule

Manage L7 rules of a loadbalancer L7 policy

### Options

```
  -h, --help   help for l7rule
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
* [ovhcloud cloud loadbalancer l7policy l7rule create](ovhcloud_cloud_loadbalancer_l7policy_l7rule_create.md)	 - Create an L7 rule in a specific L7 policy
* [ovhcloud cloud loadbalancer l7policy l7rule delete](ovhcloud_cloud_loadbalancer_l7policy_l7rule_delete.md)	 - Delete a specific L7 rule
* [ovhcloud cloud loadbalancer l7policy l7rule edit](ovhcloud_cloud_loadbalancer_l7policy_l7rule_edit.md)	 - Edit a specific L7 rule
* [ovhcloud cloud loadbalancer l7policy l7rule get](ovhcloud_cloud_loadbalancer_l7policy_l7rule_get.md)	 - Get a specific L7 rule
* [ovhcloud cloud loadbalancer l7policy l7rule list](ovhcloud_cloud_loadbalancer_l7policy_l7rule_list.md)	 - List L7 rules of a specific L7 policy

