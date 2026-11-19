## ovhcloud cloud managed-database user edit

Edit a user in the given managed database service

### Synopsis

Use this command to edit a user in the given managed database service.
There are two ways to define the edition parameters:

1. Using only CLI flags:

  For PostgreSQL and MongoDB engines:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --roles role1,role2

  For Valkey engine:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --categories "+@read" --commands "+get"

2. Using your default text editor:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --editor
  The CLI will open your default text editor to update the parameters. When saving the file, the edition will be applied.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-database user edit <service_id> <user_id> --editor --roles role1,role2


```
ovhcloud cloud managed-database user edit <service_id> <user_id> [flags]
```

### Options

```
      --categories strings   Command categories the user can execute (valkey only)
      --channels strings     Channels the user can subscribe to (valkey only)
      --commands strings     Commands the user can execute (valkey only)
      --editor               Use a text editor to define parameters
  -h, --help                 help for edit
      --keys strings         Keys the user can access (valkey only)
      --roles strings        Roles granted to the user (postgresql and mongodb only)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud managed-database user](ovhcloud_cloud_managed-database_user.md)	 - Manage users in a specific managed database service

