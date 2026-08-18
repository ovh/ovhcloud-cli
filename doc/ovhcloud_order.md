## ovhcloud order

Follow up and settle what you have ordered

### Synopsis

Follow up and settle what you have ordered.

"ovhcloud baremetal order" places an order and returns a number. These commands
are what that number is for: what state the order is in, where it is in
delivery, how to pay one placed with --no-pay, and how to give up the
retraction period — which no other command does on your behalf.

### Options

```
  -h, --help   help for order
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud order follow](ovhcloud_order_follow.md)	 - Show where an order is in delivery, step by step
* [ovhcloud order get](ovhcloud_order_get.md)	 - Retrieve one order, with the link to its order form
* [ovhcloud order list](ovhcloud_order_list.md)	 - List recent orders, their state and their price
* [ovhcloud order pay](ovhcloud_order_pay.md)	 - Pay an order that was placed without payment
* [ovhcloud order waive-retraction](ovhcloud_order_waive-retraction.md)	 - Give up the retraction period on an order

