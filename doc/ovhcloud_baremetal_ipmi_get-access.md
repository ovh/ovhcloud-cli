## ovhcloud baremetal ipmi get-access

Request an acces on KVM IPMI interface

```
ovhcloud baremetal ipmi get-access <service_name> --type serialOverLanURL --ttl 5 [flags]
```

### Options

```
      --allowed-ip string   IPv4 address that can use the access
  -h, --help                help for get-access
      --ssh-key string      Public SSH key for Serial Over Lan SSH access
      --ttl int             Time to live in minutes for cache (1, 3, 5, 10, 15) (default 1)
      --type string         Distinct way to acces a KVM IPMI session (kvmipHtml5URL, kvmipJnlp, serialOverLanSshKey, serialOverLanURL)
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
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

* [ovhcloud baremetal ipmi](ovhcloud_baremetal_ipmi.md)	 - Manage IPMI on your baremetal

