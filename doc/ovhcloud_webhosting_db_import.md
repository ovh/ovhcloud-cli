## ovhcloud webhosting db import

Import a database dump

```
ovhcloud webhosting db import <service_name> <name> [flags]
```

### Options

```
      --document-id string   Document ID from /me/documents
      --editor               Use a text editor to define parameters
      --flush                Flush database before import
      --from-file string     File containing parameters
  -h, --help                 help for import
      --send-email           Send email when done
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases

