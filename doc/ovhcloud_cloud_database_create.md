## ovhcloud cloud database create

Create a new database

### Synopsis

Use this command to create a database in the given public cloud project.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud database create --engine mysql --version 8 --plan essential  --nodes-list "db1-4:DE"

2. Using your default text editor:

	ovhcloud cloud database create --engine kafka --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud database create --engine mysql --editor --version 8


```
ovhcloud cloud database create [flags]
```

### Options

```
      --backups-regions strings          Regions on which the backups are stored
      --backups-time string              Time on which backups start every day
      --description string               Database description
      --disk-size int                    Disk size (GB)
      --editor                           Use a text editor to define parameters
      --engine string                    Database engine (you can get the list of available engines using 'ovhcloud cloud reference database list-engines')
      --fork-from.backup-id string       Backup ID (not compatible with fork-from.point-in-time)
      --fork-from.point-in-time string   Point in time to restore from (not compatible with fork-from.backup-id)
      --fork-from.service-id string      Service ID that owns the backups
  -h, --help                             help for create
      --ip-restrictions strings          IP blocks authorized to access the cluster (CIDR format)
      --maintenance-time string          Time on which maintenances can start every day
      --network-id string                Private network ID in which the cluster is deployed
      --nodes-list strings               List of nodes (format: flavor1:region1,flavor2:region2...)
      --nodes-pattern.flavor string      Flavor of all nodes
      --nodes-pattern.number int         Number of nodes
      --nodes-pattern.region string      Region of all nodes
      --plan string                      Database plan (you can get the list of available plans using 'ovhcloud cloud reference database list-plans')
      --subnet-id string                 Private subnet ID in which the cluster is deployed
      --version string                   Database version (you can get the list of available versions using 'ovhcloud cloud reference database list-engines')
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud database](ovhcloud_cloud_database.md)	 - Manage databases in the given cloud project

