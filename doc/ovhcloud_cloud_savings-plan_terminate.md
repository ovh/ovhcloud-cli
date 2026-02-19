## ovhcloud cloud savings-plan terminate

Terminate/unsubscribe from a savings plan

### Synopsis

Terminate an existing savings plan subscription.

By default, the savings plan will be terminated at the end of its current period.
You can specify a termination date using the --termination-date flag.

```
ovhcloud cloud savings-plan terminate <savings_plan_id> [flags]
```

### Options

```
  -h, --help                      help for terminate
      --termination-date string   Termination date (YYYY-MM-DD format, optional)
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

