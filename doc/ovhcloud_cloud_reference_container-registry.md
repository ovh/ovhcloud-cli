## ovhcloud cloud reference container-registry

Fetch container registry reference data in the given cloud project

### Options

```
  -h, --help   help for container-registry
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
```

### SEE ALSO

* [ovhcloud cloud reference](ovhcloud_cloud_reference.md)	 - Fetch reference data in the given cloud project
* [ovhcloud cloud reference container-registry list-plans](ovhcloud_cloud_reference_container-registry_list-plans.md)	 - List available container registry plans in the given cloud project
* [ovhcloud cloud reference container-registry list-regions](ovhcloud_cloud_reference_container-registry_list-regions.md)	 - List available container registry regions in the given cloud project

