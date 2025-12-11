## ovhcloud cloud network loadbalancer

Manage loadbalancers in the given cloud project

### Options

```
  -h, --help   help for loadbalancer
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud network](ovhcloud_cloud_network.md)	 - Manage networks in the given cloud project
* [ovhcloud cloud network loadbalancer edit](ovhcloud_cloud_network_loadbalancer_edit.md)	 - Edit the given loadbalancer
* [ovhcloud cloud network loadbalancer get](ovhcloud_cloud_network_loadbalancer_get.md)	 - Get a specific loadbalancer
* [ovhcloud cloud network loadbalancer list](ovhcloud_cloud_network_loadbalancer_list.md)	 - List your loadbalancers

