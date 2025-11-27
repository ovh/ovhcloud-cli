## ovhcloud webhosting db create

Create a database

```
  ovhcloud webhosting db create <service_name> [flags]
```

### Options

```
      --capability string   Database capability (allowed: extraSqlPerso, local, privateDatabase, sqlLocal, sqlPerso, sqlPro)
      --editor              Use a text editor to define parameters
      --from-file string    File containing parameters
  -h, --help                help for create
      --password string     Database password
      --quota string        Database quota (allowed: 25, 100, 200, 256, 400, 512, 800, 1024)
      --type string         Database type (allowed: mariadb, mysql, postgresql, redis)
      --user string         Database user (must start with hosting login, lowercase)
      --version string      Database version (allowed: 10, 10.1, 10.11, 10.2, 10.3, 10.4, 10.5, 10.6, 11, 12, 13, 15, 3.2, 3.4, 4.0, 5.1, 5.5, 5.6, 5.7, 6.0, 7.0, 8.0, 8.4, 9.4, 9.5, 9.6)
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

* [ovhcloud webhosting db](ovhcloud_webhosting_db.md)	 - Manage databases
