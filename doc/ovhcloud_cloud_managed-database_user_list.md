## ovhcloud cloud managed-database user list

List all users in the given managed database service

```
ovhcloud cloud managed-database user list <service_id> [flags]
```

### Options

```
      --filter stringArray   Filter results by any property using https://github.com/PaesslerAG/gval syntax
                             Examples:
                               --filter 'state=="running"'
                               --filter 'name=~"^my.*"'
                               --filter 'nested.property.subproperty>10'
                               --filter 'startDate>="2023-12-01"'
                               --filter 'name=~"something" && nbField>10'
  -h, --help                 help for list
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

* [ovhcloud cloud managed-database user](ovhcloud_cloud_managed-database_user.md)	 - Manage users in a specific managed database service

