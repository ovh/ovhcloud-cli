## ovhcloud cloud savings-plan set-renewal

Set the action at the end of the savings plan period

### Synopsis

Set the action to be performed when the savings plan reaches the end of its period.

Available actions:
- REACTIVATE: Automatically renew the savings plan for another period
- TERMINATE: Terminate the savings plan at the end of the period

```
ovhcloud cloud savings-plan set-renewal <savings_plan_id> [flags]
```

### Options

```
      --action string   Action at period end: REACTIVATE or TERMINATE (required)
  -h, --help            help for set-renewal
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
```

### SEE ALSO

* [ovhcloud cloud savings-plan](ovhcloud_cloud_savings-plan.md)	 - Manage savings plans for your cloud project

