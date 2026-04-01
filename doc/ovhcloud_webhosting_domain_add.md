## ovhcloud webhosting domain add

Attach a domain

```
ovhcloud webhosting domain add <service_name> [flags]
```

### Options

```
      --bypass-dns           If set to true, DNS zone will not be updated by the operation
      --cdn string           Whether the attached domain is linked to the hosting CDN (allowed: active, none)
      --disable-ssl          Exclude the attached domain from the SSL certificate
      --domain string        Domain to link
      --editor               Use a text editor to define parameters
      --enable-ssl           Whether to put the attached domain in the SSL certificate
      --firewall string      Whether the firewall is active for this domain (allowed: active, none)
      --from-file string     File containing parameters
  -h, --help                 help for add
      --ip-location string   Change attached domain's DNS to the IP of the country (allowed: BE, CA, CZ, DE, ES, FI, FR, IE, IT, LT, NL, PL, PT, UK)
      --own-log string       Domain to separate the logs on
      --path string          Path of the attached domain
      --runtime-id int       Runtime configuration ID used on this domain
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

* [ovhcloud webhosting domain](ovhcloud_webhosting_domain.md)	 - Manage attached domains

