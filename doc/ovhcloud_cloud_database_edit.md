## ovhcloud cloud database edit

Edit a specific database

### Synopsis

Use this command to edit a database in the given public cloud project.
There are two ways to define the edition parameters:

1. Using only CLI flags:

	ovhcloud cloud database edit <database_id> --description "My database"

2. Using your default text editor:

	ovhcloud cloud database edit <database_id> --editor

  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud database edit <database_id> --editor --description "My database"


```
ovhcloud cloud database edit <database_id> [flags]
```

### Options

```
      --backups-regions strings   Regions on which the backups are stored
      --backups-time string       Time on which backups start every day
      --deletion-protection       Enable deletion protection
      --description string        Description of the cluster
      --editor                    Use a text editor to define parameters
      --enable-prometheus         Enable Prometheus
      --flavor string             The VM flavor used for this cluster
  -h, --help                      help for edit
      --ip-restrictions strings   IP blocks authorized to access the cluster (CIDR format)
      --maintenance-time string   Time on which maintenances can start every day
      --plan string               Plan of the cluster
      --version string            Version of the engine deployed on the cluster
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

* [ovhcloud cloud database](ovhcloud_cloud_database.md)	 - Manage databases in the given cloud project

