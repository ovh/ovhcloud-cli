## ovhcloud vps disk edit

Edit a specific disk of the given VPS

```
ovhcloud vps disk edit <service_name> <disk_id> [flags]
```

### Options

```
      --editor                         Use a text editor to define parameters
  -h, --help                           help for edit
      --low-free-space-threshold int   Low free space threshold for the disk
      --monitoring                     Enable or disable monitoring for the disk
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

* [ovhcloud vps disk](ovhcloud_vps_disk.md)	 - Manage disks of the given VPS

