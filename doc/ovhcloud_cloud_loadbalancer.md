## ovhcloud cloud loadbalancer

Manage loadbalancers in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for loadbalancer
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
* [ovhcloud cloud loadbalancer associate-floating-ip](ovhcloud_cloud_loadbalancer_associate-floating-ip.md)	 - Associate an existing floating IP to a loadbalancer
* [ovhcloud cloud loadbalancer create](ovhcloud_cloud_loadbalancer_create.md)	 - Create a loadbalancer in the given cloud project
* [ovhcloud cloud loadbalancer create-floating-ip](ovhcloud_cloud_loadbalancer_create-floating-ip.md)	 - Create a floating IP and attach it to a loadbalancer
* [ovhcloud cloud loadbalancer delete](ovhcloud_cloud_loadbalancer_delete.md)	 - Delete a specific loadbalancer
* [ovhcloud cloud loadbalancer edit](ovhcloud_cloud_loadbalancer_edit.md)	 - Edit the given loadbalancer
* [ovhcloud cloud loadbalancer get](ovhcloud_cloud_loadbalancer_get.md)	 - Get a specific loadbalancer
* [ovhcloud cloud loadbalancer get-flavor](ovhcloud_cloud_loadbalancer_get-flavor.md)	 - Get details of a specific loadbalancer flavor
* [ovhcloud cloud loadbalancer health-monitor](ovhcloud_cloud_loadbalancer_health-monitor.md)	 - Manage health monitors of loadbalancers
* [ovhcloud cloud loadbalancer l7policy](ovhcloud_cloud_loadbalancer_l7policy.md)	 - Manage L7 policies of loadbalancers
* [ovhcloud cloud loadbalancer list](ovhcloud_cloud_loadbalancer_list.md)	 - List your loadbalancers
* [ovhcloud cloud loadbalancer list-flavors](ovhcloud_cloud_loadbalancer_list-flavors.md)	 - List available loadbalancer flavors in the given cloud project
* [ovhcloud cloud loadbalancer listener](ovhcloud_cloud_loadbalancer_listener.md)	 - Manage listeners of loadbalancers
* [ovhcloud cloud loadbalancer log](ovhcloud_cloud_loadbalancer_log.md)	 - Manage loadbalancer logs
* [ovhcloud cloud loadbalancer pool](ovhcloud_cloud_loadbalancer_pool.md)	 - Manage pools of loadbalancers
* [ovhcloud cloud loadbalancer stats](ovhcloud_cloud_loadbalancer_stats.md)	 - Get statistics for a loadbalancer

