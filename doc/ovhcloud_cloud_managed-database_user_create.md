## ovhcloud cloud managed-database user create

Create a new user in the given managed database service

### Synopsis

Use this command to create a user in the given managed database service.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-database user create <service_id> --name myuser

  For PostgreSQL and MongoDB engines, you can also specify roles:

	ovhcloud cloud managed-database user create <service_id> --name myuser --roles role1,role2

  For Valkey engine, you can specify permissions:

	ovhcloud cloud managed-database user create <service_id> --name myuser --categories "+@read" --commands "+get"

2. Using your default text editor:

	ovhcloud cloud managed-database user create <service_id> --name myuser --editor
  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-database user create <service_id> --name myuser --editor --roles role1,role2


```
ovhcloud cloud managed-database user create <service_id> [flags]
```

### Options

```
      --categories strings   Command categories the user can execute (valkey only)
      --channels strings     Channels the user can subscribe to (valkey only)
      --commands strings     Commands the user can execute (valkey only)
      --editor               Use a text editor to define parameters
  -h, --help                 help for create
      --keys strings         Keys the user can access (valkey only)
      --name string          Name of the user to create
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

