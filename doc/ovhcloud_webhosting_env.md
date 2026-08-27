## ovhcloud webhosting env

Manage environment variables

### Options

```
  -h, --help   help for env
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting env create](ovhcloud_webhosting_env_create.md)	 - Create an env var
* [ovhcloud webhosting env delete](ovhcloud_webhosting_env_delete.md)	 - Delete an env var
* [ovhcloud webhosting env get](ovhcloud_webhosting_env_get.md)	 - Get an env var
* [ovhcloud webhosting env list](ovhcloud_webhosting_env_list.md)	 - List env vars
* [ovhcloud webhosting env update](ovhcloud_webhosting_env_update.md)	 - Update an env var

