## ovhcloud baremetal catalog

List orderable servers, their availability and their price

### Synopsis

List orderable servers, where they can be delivered, how long delivery takes
and what they cost.

Delivery is a delay, not a yes or no: "24h" means a server ordered now is
expected within a day, while "2mo" is a two-month wait. Rows are sorted with the
soonest first.

Scale, High Grade and SAP servers appear with their availability but no price:
they are not sold from the public price list, and show as "on quotation".

```
ovhcloud baremetal catalog [flags]
```

### Options

```
      --available-only          Hide what cannot be delivered today
      --commitment string       Which price to show: default (monthly), 12 or 24 (months paid upfront) (default "default")
      --country string          Subsidiary whose price list to read (default: the one this account belongs to)
      --datacenter strings      Only these datacenters (repeatable)
      --filter stringArray      Filter results by any property using https://github.com/PaesslerAG/gval syntax
                                Examples:
                                  --filter 'state=="running"'
                                  --filter 'name=~"^my.*"'
                                  --filter 'nested.property.subproperty>10'
                                  --filter 'startDate>="2023-12-01"'
                                  --filter 'name=~"something" && nbField>10'
      --gpu string              Only this GPU reference
  -h, --help                    help for catalog
      --memory string           Only this memory reference
      --plan-code string        Only this plan code (the identifier used to order)
      --refresh                 Download the price list again instead of reusing the one cached today
      --region strings          Only these regions, reported per region instead of per datacenter (repeatable)
      --server string           Only this base hardware, for example 24ska01
      --storage string          Only this storage reference
      --system-storage string   Only this system storage reference
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

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services

