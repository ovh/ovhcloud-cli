## ovhcloud order waive-retraction

Give up the retraction period on an order

### Synopsis

Give up the retraction period on an order.

This cannot be undone. It is a command of its own so that it is never a side
effect of buying: "ovhcloud baremetal order" will not waive the retraction
period even though the order API accepts the field, because giving up a legal
right belongs in a sentence you typed on purpose.

```
ovhcloud order waive-retraction <order_id> [flags]
```

### Options

```
      --dry-run   Print the call that would be made without making it
  -h, --help      help for waive-retraction
  -y, --yes       Skip the confirmation prompt (required for unattended runs)
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud order](ovhcloud_order.md)	 - Follow up and settle what you have ordered

