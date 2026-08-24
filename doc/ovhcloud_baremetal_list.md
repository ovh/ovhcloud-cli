## ovhcloud baremetal list

List your Baremetal services

### Synopsis

List your Baremetal services.

--tag narrows the list on the API itself, using the IAM tags set on the servers. Write it as key=value, or key:OPERATOR=value to compare another way; OPERATOR is one of the names the API uses (EQ, NEQ, LIKE, ILIKE, EXISTS, NEXISTS), and EXISTS and NEXISTS take no value. Several --tag narrow further.

It is not the same thing as --filter, which runs on the columns of the table once the servers have been read.

```
ovhcloud baremetal list [flags]
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
      --tag stringArray      Only list servers carrying this IAM tag (key=value, or key:OPERATOR=value)
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

