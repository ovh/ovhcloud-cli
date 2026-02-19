## ovhcloud cloud container-registry iam enable

Enable IAM for the given container registry

```
ovhcloud cloud container-registry iam enable <registry_id> [flags]
```

### Options

```
      --delete-users       Delete existing container registry users when enabling IAM
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for enable
      --init-file string   Create a file with example parameters
      --replace            Replace parameters file if it already exists
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
```

### SEE ALSO

* [ovhcloud cloud container-registry iam](ovhcloud_cloud_container-registry_iam.md)	 - Manage container registry IAM

