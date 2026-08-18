## ovhcloud baremetal order

Order a dedicated server

### Synopsis

Order one of the plan codes listed by "ovhcloud baremetal catalog".

This command spends money. It builds a cart, prices it, shows what the order
will cost and what has to be accepted, and asks for the datacenter to be typed
back before it buys anything. --quote stops at the price and never orders;
--dry-run describes the whole sequence and sends nothing at all.

The server is delivered without an operating system: a dedicated server is
installed after delivery with "ovhcloud baremetal reinstall", and the order
carries no choice of system. The retraction period is never waived here.

```
ovhcloud baremetal order <plan_code> [flags]
```

### Examples

```
  # What would this cost, and what would it need?
  ovhcloud baremetal order 24adv01-v3 --datacenter gra --quote

  # Buy it, monthly
  ovhcloud baremetal order 24adv01-v3 --datacenter gra

  # Buy two on a twelve-month commitment, unattended
  ovhcloud baremetal order 24adv01-v3 --datacenter gra --commitment 12 --quantity 2 --yes
```

### Options

```
      --commitment string    How it is paid: default (monthly), 12 or 24 (months paid upfront) (default "default")
      --config stringArray   A configuration this product requires, as label=value (repeatable)
      --datacenter string    Where to deliver the server, for example gra
      --dry-run              Describe the whole ordering sequence without sending any of it
  -h, --help                 help for order
      --no-pay               Place the order without paying it, and return the order number to pay later
      --quantity int         How many servers to order (default 1)
      --quote                Stop at the price: build the cart, show what it would cost, order nothing
  -y, --yes                  Skip the confirmation prompt (required for unattended runs)
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

