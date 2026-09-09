## ovhcloud api call

Call any OVHcloud API path

### Synopsis

Call any OVHcloud API path with the credentials already configured.

The path may be given with or without its leading slash and API version:
"dedicated/server", "/dedicated/server" and "/v1/dedicated/server" are the
same endpoint. A path already starting with /v1/ or /v2/ is left untouched.

A body for POST and PUT is read from --from-file or written in $EDITOR with
--editor.

```
ovhcloud api call <method> <path> [flags]
```

### Examples

```
  # List the dedicated servers of the account
  ovhcloud api call GET /dedicated/server

  # Read one server, as JSON
  ovhcloud api call GET /dedicated/server/ns3168421.ip-51-77-12.eu -o json

  # Extract a single field
  ovhcloud api call GET /dedicated/server/ns3168421.ip-51-77-12.eu -o 'datacenter'

  # Reach a v2 endpoint
  ovhcloud api call GET /v2/publicCloud/project

  # Send a body, checking first what would leave
  ovhcloud api call PUT /dedicated/server/ns3168421.ip-51-77-12.eu --from-file body.json --dry-run
```

### Options

```
      --dry-run            Print the request that would be sent, without sending it
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for call
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

* [ovhcloud api](ovhcloud_api.md)	 - Call the OVHcloud API directly

