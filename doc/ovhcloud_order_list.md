## ovhcloud order list

List recent orders, their state and their price

### Synopsis

List recent orders, their state and their price.

The state of an order is a separate call from the order itself, so this costs
two requests per order. That is why the window is thirty days and the number of
orders is capped by default, and why both bounds are printed when they bite
rather than applied in silence.

```
ovhcloud order list [flags]
```

### Options

```
      --days int             How far back to look (default 30)
      --filter stringArray   Filter results by any property using https://github.com/PaesslerAG/gval syntax
                             Examples:
                               --filter 'state=="running"'
                               --filter 'name=~"^my.*"'
                               --filter 'nested.property.subproperty>10'
                               --filter 'startDate>="2023-12-01"'
                               --filter 'name=~"something" && nbField>10'
      --from string          Start of the window, as 2026-08-01T00:00:00Z (instead of --days)
  -h, --help                 help for list
      --limit int            How many of the most recent orders to show (0 for all of them) (default 25)
      --to string            End of the window, as 2026-08-31T00:00:00Z
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

