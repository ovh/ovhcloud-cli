## ovhcloud cloud container-registry ip-restrictions management add

Add a management IP restriction to a container registry

```
ovhcloud cloud container-registry ip-restrictions management add <registry_id> [flags]
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

* [ovhcloud cloud container-registry ip-restrictions management](ovhcloud_cloud_container-registry_ip-restrictions_management.md)	 - Manage IP restrictions for container registry Harbor UI and API access

