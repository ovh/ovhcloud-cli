## ovhcloud cloud key-manager secret consumer

Manage consumers of a Key Manager secret

### Options

```
  -h, --help   help for consumer
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
      --raw                    Output the extracted value without JSON quoting (use with -o '<field>'), useful for scripting
                               Example:
                                 --output 'id' --raw   (prints the id without surrounding quotes)
```

### SEE ALSO

* [ovhcloud cloud key-manager secret](ovhcloud_cloud_key-manager_secret.md)	 - Manage Key Manager secrets
* [ovhcloud cloud key-manager secret consumer delete](ovhcloud_cloud_key-manager_secret_consumer_delete.md)	 - Delete a consumer from the given secret
* [ovhcloud cloud key-manager secret consumer get](ovhcloud_cloud_key-manager_secret_consumer_get.md)	 - Get a specific consumer of the given secret
* [ovhcloud cloud key-manager secret consumer list](ovhcloud_cloud_key-manager_secret_consumer_list.md)	 - List consumers registered for the given secret
* [ovhcloud cloud key-manager secret consumer register](ovhcloud_cloud_key-manager_secret_consumer_register.md)	 - Register a consumer for the given secret

