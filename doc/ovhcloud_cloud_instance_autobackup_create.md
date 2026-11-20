## ovhcloud cloud instance autobackup create

Create an automatic backup workflow for the given instance

### Synopsis

Create an automatic backup workflow for the given instance, in the region of the instance.

The parameters can be given with flags, with a parameters file (--from-file, see --init-file
to generate an example) or with your text editor (--editor). Flags override the values of the file.

The instance and its region come from <instance_id>: they override the values of a parameters
file, and are pre-filled in the editor, where they should be left unchanged.

With flags only, --cron, --rotation and --name are required, and the backup images are named after
the workflow (--name) unless --image-name is given. A parameters file must give the image name.

```
ovhcloud cloud instance autobackup create <instance_id> [flags]
```

### Options

```
      --cron string                 Unix Cron pattern (e.g. '0 0 * * *')
      --distant-image-name string   Name of the backup images copied to the distant region
      --distant-region string       Region to copy each backup to (cross-region backup)
      --editor                      Use a text editor to define parameters
      --from-file string            File containing parameters
  -h, --help                        help for create
      --image-name string           Name prefix of the backup images (with flags only, defaults to the workflow name)
      --init-file string            Create a file with example parameters
      --name string                 Workflow name
      --replace                     Replace parameters file if it already exists
      --rotation int                Number of backups to keep
      --wait                        Wait for the backup workflow to be ready before exiting
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
```

### SEE ALSO

* [ovhcloud cloud instance autobackup](ovhcloud_cloud_instance_autobackup.md)	 - Manage automatic backup workflows for instances

