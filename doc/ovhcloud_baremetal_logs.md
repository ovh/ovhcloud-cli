## ovhcloud baremetal logs

Read the logs of a dedicated server, and send them to a stream

### Options

```
  -h, --help   help for logs
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
* [ovhcloud baremetal logs kinds](ovhcloud_baremetal_logs_kinds.md)	 - List the kinds of log this server can send
* [ovhcloud baremetal logs subscribe](ovhcloud_baremetal_logs_subscribe.md)	 - Send the logs of this server to a Log Data Platform stream
* [ovhcloud baremetal logs subscription](ovhcloud_baremetal_logs_subscription.md)	 - Show where the logs of this server are sent
* [ovhcloud baremetal logs unsubscribe](ovhcloud_baremetal_logs_unsubscribe.md)	 - Stop sending the logs of this server to a stream
* [ovhcloud baremetal logs url](ovhcloud_baremetal_logs_url.md)	 - Get a temporary link to read the logs of this server

