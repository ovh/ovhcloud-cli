## ovhcloud baremetal logs subscribe

Send the logs of this server to a Log Data Platform stream

### Synopsis

Send the logs of this server to a Log Data Platform stream.

--stream takes the title of a stream or its identifier. A title is resolved across every Log Data Platform service on the account, and a title carried by more than one stream is refused rather than guessed.

```
ovhcloud baremetal logs subscribe <service_name> [flags]
```

### Options

```
      --dry-run         Print the call that would be made without making it
  -h, --help            help for subscribe
      --kind string     Kind of log to work on (default: the only one the server offers)
      --stream string   Title or identifier of the stream to send the logs to
      --wait            Wait until the subscription actually exists before exiting
  -y, --yes             Skip the confirmation prompt (required for unattended runs)
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

