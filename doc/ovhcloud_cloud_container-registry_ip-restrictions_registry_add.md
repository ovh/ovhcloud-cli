## ovhcloud cloud container-registry ip-restrictions registry add

Add a registry IP restriction to a container registry

```
ovhcloud cloud container-registry ip-restrictions registry add <registry_id> [flags]
```

### Options

```
      --description string   Description for the IP restriction (optional)
  -h, --help                 help for add
      --ip-block string      IP block in CIDR notation (e.g., 192.0.2.0/24)
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

* [ovhcloud cloud container-registry ip-restrictions registry](ovhcloud_cloud_container-registry_ip-restrictions_registry.md)	 - Manage IP restrictions for container registry artifact manager (Docker, Helm...) access

