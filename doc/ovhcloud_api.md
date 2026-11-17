## ovhcloud api

Call the OVHcloud API directly

### Synopsis

Call any OVHcloud API endpoint with the credentials already configured.

The CLI covers a part of the API surface: this is the way out for the rest.
Requests are signed with the active profile and rendered with the usual
--output formats, so an endpoint with no dedicated command is still usable
from a script.

It is a raw passthrough. None of the confirmations, parameter validation or
safety checks the product commands provide apply here, so use --dry-run to
see what would be sent before sending it.

### Options

```
  -h, --help   help for api
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
* [ovhcloud api call](ovhcloud_api_call.md)	 - Call any OVHcloud API path

