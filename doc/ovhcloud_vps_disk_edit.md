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
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud vps disk](ovhcloud_vps_disk.md)	 - Manage disks of the given VPS

