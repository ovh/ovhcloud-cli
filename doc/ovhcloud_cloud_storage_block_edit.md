## ovhcloud cloud storage block edit

Edit a volume (prompts for one if omitted)

```
ovhcloud cloud storage block edit [volume_id] [flags]
```

### Examples

```
  # Rename a volume and grow it to 40 GB
  ovhcloud cloud storage block edit <volume_id> --cloud-project <project_id> --name backups --size 40
```

### Options

```
      --description string   Volume description
      --editor               Use a text editor to define parameters
  -h, --help                 help for edit
      --name string          Volume name
      --size int             Volume size (in GB, can only be increased)
      --type string          Volume type (classic, classic-luks, classic-multiattach, high-speed, high-speed-gen2, high-speed-gen2-luks, high-speed-luks)
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
  -y, --yes                    Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud cloud storage block](ovhcloud_cloud_storage_block.md)	 - Manage block storage volumes in the given cloud project

