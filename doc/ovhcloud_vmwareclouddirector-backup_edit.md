## ovhcloud vmwareclouddirector-backup edit

Edit the given VmwareCloudDirector Backup

```
ovhcloud vmwareclouddirector-backup edit <service_name> [flags]
```

### Options

```
      --editor           Use a text editor to define parameters
  -h, --help             help for edit
      --offers strings   List of your VMware Cloud Director backup offers formatted as '<name>:<quotaInTB>' (available names: BRONZE, GOLD, SILVER)
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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

* [ovhcloud vmwareclouddirector-backup](ovhcloud_vmwareclouddirector-backup.md)	 - Retrieve information and manage your VMware Cloud Director Backup services

