## ovhcloud cloud ip

Manage public IPs (floating and failover) in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for ip
      --region string          Filter by region or specify the region of the floating IP (only used when --type=floating)
      --type string            Type of IP to manage (floating or failover)
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
  -y, --yes              Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud ip attach](ovhcloud_cloud_ip_attach.md)	 - Attach a public IP to an instance (only supported for --type=failover)
* [ovhcloud cloud ip delete](ovhcloud_cloud_ip_delete.md)	 - Delete a public IP (only supported for --type=floating)
* [ovhcloud cloud ip get](ovhcloud_cloud_ip_get.md)	 - Get information about a public IP
* [ovhcloud cloud ip list](ovhcloud_cloud_ip_list.md)	 - List public IPs (both floating and failover when --type is not specified)

