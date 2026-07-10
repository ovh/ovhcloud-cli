## ovhcloud cloud ip failover

Manage failover public IPs in the given cloud project

### Options

```
  -h, --help   help for failover
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud ip](ovhcloud_cloud_ip.md)	 - Manage public IPs (floating, additional, ext-net and failover) in the given cloud project
* [ovhcloud cloud ip failover attach](ovhcloud_cloud_ip_failover_attach.md)	 - Attach a failover IP to an instance
* [ovhcloud cloud ip failover get](ovhcloud_cloud_ip_failover_get.md)	 - Get a specific failover IP
* [ovhcloud cloud ip failover list](ovhcloud_cloud_ip_failover_list.md)	 - List failover IPs

