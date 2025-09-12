## ovhcloud domain-name edit

Edit the given domain name service

```
ovhcloud domain-name edit <domain_name> [flags]
```

### Options

```
      --editor                        Use a text editor to define parameters
  -h, --help                          help for edit
      --name-server-type string       Type of name server (anycast, dedicated, empty, external, hold, hosted, hosting, mixed, parking)
      --transfer-lock-status string   Transfer lock status (locked, locking, unavailable, unlocked, unlocking)
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using gval format)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud domain-name](ovhcloud_domain-name.md)	 - Retrieve information and manage your domain names

