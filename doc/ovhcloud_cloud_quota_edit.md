## ovhcloud cloud quota edit

Update the project quota (target quota profile per region)

```
ovhcloud cloud quota edit [flags]
```

### Examples

```
  # Edit the project quota interactively (choose the target profile per region)
  ovhcloud cloud quota edit --cloud-project <project_id> --editor
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for edit
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud quota](ovhcloud_cloud_quota.md)	 - Manage quotas in the given cloud project

