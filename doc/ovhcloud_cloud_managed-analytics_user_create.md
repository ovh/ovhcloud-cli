## ovhcloud cloud managed-analytics user create

Create a new user in the given managed analytics service

### Synopsis

Use this command to create a user in the given managed analytics service.
There are two ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser

  For Clickhouse engine, you can also specify roles:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --roles role1,role2

  For OpenSearch engine, you can specify ACLs:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --acls "logs-.*:write" --acls "metrics-.*:read"

2. Using your default text editor:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --editor
  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud managed-analytics user create <service_id> --name myuser --editor --roles role1,role2


```
ovhcloud cloud managed-analytics user create <service_id> [flags]
```

### Options

```
      --acls stringArray   ACL granted to the user (opensearch only, format: pattern:permission)
      --editor             Use a text editor to define parameters
  -h, --help               help for create
      --name string        Name of the user to create
      --roles strings      Roles granted to the user (clickhouse only)
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

* [ovhcloud cloud managed-analytics user](ovhcloud_cloud_managed-analytics_user.md)	 - Manage users in a specific managed analytics service

