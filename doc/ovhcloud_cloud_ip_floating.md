## ovhcloud cloud ip floating

Manage floating public IPs in the given cloud project

### Options

```
  -h, --help   help for floating
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

* [ovhcloud cloud ip](ovhcloud_cloud_ip.md)	 - Manage public IPs (floating, additional and ext-net) in the given cloud project
* [ovhcloud cloud ip floating create](ovhcloud_cloud_ip_floating_create.md)	 - Create a new floating IP
* [ovhcloud cloud ip floating delete](ovhcloud_cloud_ip_floating_delete.md)	 - Delete a specific floating IP
* [ovhcloud cloud ip floating edit](ovhcloud_cloud_ip_floating_edit.md)	 - Edit the given floating IP
* [ovhcloud cloud ip floating get](ovhcloud_cloud_ip_floating_get.md)	 - Get a specific floating IP
* [ovhcloud cloud ip floating list](ovhcloud_cloud_ip_floating_list.md)	 - List floating IPs

