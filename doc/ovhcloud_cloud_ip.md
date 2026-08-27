## ovhcloud cloud ip

Manage public IPs (floating, additional and ext-net) in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for ip
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
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud ip additional](ovhcloud_cloud_ip_additional.md)	 - Manage additional public IPs in the given cloud project
* [ovhcloud cloud ip extnet](ovhcloud_cloud_ip_extnet.md)	 - Manage ext-net public IPs in the given cloud project
* [ovhcloud cloud ip floating](ovhcloud_cloud_ip_floating.md)	 - Manage floating public IPs in the given cloud project
* [ovhcloud cloud ip list](ovhcloud_cloud_ip_list.md)	 - List all public IPs (floating, additional and ext-net) of the project

