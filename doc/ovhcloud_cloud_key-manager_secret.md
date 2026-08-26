## ovhcloud cloud key-manager secret

Manage Key Manager secrets

### Options

```
  -h, --help   help for secret
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

* [ovhcloud cloud key-manager](ovhcloud_cloud_key-manager.md)	 - Manage Key Management Service (KMS) resources in the given cloud project
* [ovhcloud cloud key-manager secret consumer](ovhcloud_cloud_key-manager_secret_consumer.md)	 - Manage consumers of a Key Manager secret
* [ovhcloud cloud key-manager secret create](ovhcloud_cloud_key-manager_secret_create.md)	 - Create a new Key Manager secret
* [ovhcloud cloud key-manager secret delete](ovhcloud_cloud_key-manager_secret_delete.md)	 - Delete the given Key Manager secret
* [ovhcloud cloud key-manager secret edit](ovhcloud_cloud_key-manager_secret_edit.md)	 - Edit the given Key Manager secret (only metadata is mutable)
* [ovhcloud cloud key-manager secret get](ovhcloud_cloud_key-manager_secret_get.md)	 - Get a specific Key Manager secret
* [ovhcloud cloud key-manager secret list](ovhcloud_cloud_key-manager_secret_list.md)	 - List Key Manager secrets
* [ovhcloud cloud key-manager secret payload](ovhcloud_cloud_key-manager_secret_payload.md)	 - Manage the payload (sensitive material) of a Key Manager secret

