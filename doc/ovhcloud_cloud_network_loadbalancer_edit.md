## ovhcloud cloud network loadbalancer edit

Edit the given loadbalancer

```
ovhcloud cloud network loadbalancer edit <loadbalancer_id> [flags]
```

### Options

```
      --description string   Description of the loadbalancer
      --editor               Use a text editor to define parameters
      --flavor string        Flavor ID of the loadbalancer (can be retrieved with 'cloud reference loadbalancer list-flavors <region>')
  -h, --help                 help for edit
      --name string          Name of the loadbalancer
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

* [ovhcloud cloud network loadbalancer](ovhcloud_cloud_network_loadbalancer.md)	 - Manage loadbalancers in the given cloud project

