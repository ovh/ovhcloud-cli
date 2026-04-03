## ovhcloud cloud floating-ip

Manage floating IPs in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for floating-ip
      --region string          Filter by region or specify the region of the floating IP
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud floating-ip delete](ovhcloud_cloud_floating-ip_delete.md)	 - Delete a floating IP
* [ovhcloud cloud floating-ip get](ovhcloud_cloud_floating-ip_get.md)	 - Get information about a floating IP
* [ovhcloud cloud floating-ip list](ovhcloud_cloud_floating-ip_list.md)	 - List floating IPs

