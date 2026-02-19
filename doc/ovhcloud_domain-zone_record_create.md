## ovhcloud domain-zone record create

Create a single DNS record in your zone

```
ovhcloud domain-zone record create <zone_name> [flags]
```

### Options

```
      --editor              Use a text editor to define parameters
      --field-type string   Record type (A, AAAA, CAA, CNAME, DKIM, DMARC, DNAME, HTTPS, LOC, MX, NAPTR, NS, PTR, RP, SPF, SRV, SSHFP, SVCB, TLSA, TXT)
      --from-file string    File containing parameters
  -h, --help                help for create
      --init-file string    Create a file with example parameters
      --replace             Replace parameters file if it already exists
      --sub-domain string   Record subDomain
      --target string       Target of the record
      --ttl int             TTL of the record
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
```

### SEE ALSO

* [ovhcloud domain-zone record](ovhcloud_domain-zone_record.md)	 - Retrieve information and manage your DNS records within a zone

