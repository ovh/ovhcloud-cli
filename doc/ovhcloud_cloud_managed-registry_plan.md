## ovhcloud cloud managed-registry plan

Manage container registry plans

### Options

```
  -h, --help   help for plan
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
      --raw                    Output the extracted value without JSON quoting (use with -o '<field>'), useful for scripting
                               Example:
                                 --output 'id' --raw   (prints the id without surrounding quotes)
```

### SEE ALSO

* [ovhcloud cloud managed-registry](ovhcloud_cloud_managed-registry.md)	 - Manage container registries in the given cloud project
* [ovhcloud cloud managed-registry plan list](ovhcloud_cloud_managed-registry_plan_list.md)	 - List available container registry plans in the given cloud project
* [ovhcloud cloud managed-registry plan list-capabilities](ovhcloud_cloud_managed-registry_plan_list-capabilities.md)	 - List available plans for a specific container registry
* [ovhcloud cloud managed-registry plan upgrade](ovhcloud_cloud_managed-registry_plan_upgrade.md)	 - Upgrade a container registry plan

