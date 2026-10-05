## ovhcloud domain-zone record update

Update a single DNS record from your zone

```
ovhcloud domain-zone record update <zone_name> <record_id> [flags]
```

### Options

```
      --editor              Use a text editor to define parameters
      --from-file string    File containing parameters
  -h, --help                help for update
      --init-file string    Create a file with example parameters
      --replace             Replace parameters file if it already exists
      --sub-domain string   Subdomain to update
      --target string       New target to apply
      --ttl int             New TTL to apply
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud domain-zone record](ovhcloud_domain-zone_record.md)	 - Retrieve information and manage your DNS records within a zone

