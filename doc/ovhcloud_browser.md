## ovhcloud browser

Launch a TUI for the OVHcloud Manager - Public Cloud universe only [EXPERIMENTAL]

### Synopsis

Launch an interactive Terminal User Interface that simulates the
OVHcloud Manager (https://manager.eu.ovhcloud.com/#/public-cloud/) - Public Cloud universe only.

⚠️  EXPERIMENTAL FEATURE - This navigation is experimental and may contain bugs.
If you encounter any issues, please report them at:
https://github.com/ovh/ovhcloud-cli/issues

Navigate through your Public Cloud services using keyboard controls.
The browser makes direct API calls to fetch and display real data.

Features:
  - Real-time data fetching from OVHcloud API
  - Table views for projects, instances, and services
  - Hierarchical navigation through cloud resources
  - Web-like interface in your terminal
  - Debug mode to view API requests and request IDs

Navigate using:
  - ↑↓: Move through menus/tables
  - Enter: Select item or view details
  - ←/Esc: Go back
  - d: Toggle debug panel (show API requests)
  - q: Quit

```
ovhcloud browser [flags]
```

### Options

```
      --debug   Enable debug mode to view API requests and request IDs
  -h, --help    help for browser
```

### Options inherited from parent commands

```
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services

