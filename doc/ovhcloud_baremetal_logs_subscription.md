## ovhcloud baremetal logs subscription

Show where the logs of this server are sent

### Options

```
  -h, --help   help for subscription
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

* [ovhcloud baremetal logs](ovhcloud_baremetal_logs.md)	 - Read the logs of a dedicated server, and send them to a stream
* [ovhcloud baremetal logs subscription get](ovhcloud_baremetal_logs_subscription_get.md)	 - Show one log subscription of this server
* [ovhcloud baremetal logs subscription list](ovhcloud_baremetal_logs_subscription_list.md)	 - List the log subscriptions of this server

