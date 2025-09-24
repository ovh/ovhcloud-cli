## ovhcloud vps disk

Manage disks of the given VPS

### Options

```
  -h, --help   help for disk
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud vps](ovhcloud_vps.md)	 - Retrieve information and manage your VPS services
* [ovhcloud vps disk edit](ovhcloud_vps_disk_edit.md)	 - Edit a specific disk of the given VPS
* [ovhcloud vps disk get](ovhcloud_vps_disk_get.md)	 - Get information about a specific disk of the given VPS
* [ovhcloud vps disk list](ovhcloud_vps_disk_list.md)	 - List disks of the given VPS

