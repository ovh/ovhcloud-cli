## ovhcloud baremetal doctor

Report what is wrong with a server, or with every server

### Synopsis

Check the things that silently break a dedicated server: a machine left on the
rescue system, monitoring switched off, hardware intervention refused, a renewal
that will not happen, work still running, maintenance already planned.

With no argument it checks every server of the account.

The exit code stays 0 when findings are reported, because the command ran and
answered. Use --strict to make findings fail the command instead, which is what
a pipeline gating on it wants.

--strict fails on a warning or a critical. A note never fails it: every server
renewing inside the next 30 days reports one, so a --strict that counted notes
would be red permanently, and a gate that is always red is read like no gate.
Narrowing with --filter narrows the gate too — 'severity=="critical"' fails only
on criticals.

```
ovhcloud baremetal doctor [service_name...] [flags]
```

### Options

```
      --expiry-days int      Report a server expiring within this many days (default 30)
      --filter stringArray   Filter results by any property using https://github.com/PaesslerAG/gval syntax
                             Examples:
                               --filter 'state=="running"'
                               --filter 'name=~"^my.*"'
                               --filter 'nested.property.subproperty>10'
                               --filter 'startDate>="2023-12-01"'
                               --filter 'name=~"something" && nbField>10'
  -h, --help                 help for doctor
      --strict               Exit non-zero when a warning or a critical is reported (notes never fail it)
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

