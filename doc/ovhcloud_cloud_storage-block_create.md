## ovhcloud cloud storage-block create

Create a new volume

```
ovhcloud cloud storage-block create <region> [flags]
```

### Options

```
      --availability-zone string   Availability zone of the volume
      --backup-id string           Backup ID
      --description string         Volume description
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --image-id string            Image ID to create the volume from
      --init-file string           Create a file with example parameters
      --instance-id string         Instance ID to attach the volume to
      --name string                Volume name
      --replace                    Replace parameters file if it already exists
      --size int                   Volume size (in GB)
      --snapshot-id string         Snapshot ID to create the volume from
      --type string                Volume type (classic, classic-luks, classic-multiattach, high-speed, high-speed-gen2, high-speed-gen2-luks, high-speed-luks)
      --wait                       Wait for volume creation to be done before exiting
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

* [ovhcloud cloud storage-block](ovhcloud_cloud_storage-block.md)	 - Manage block storage volumes in the given cloud project

