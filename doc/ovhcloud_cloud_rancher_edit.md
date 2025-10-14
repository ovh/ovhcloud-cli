## ovhcloud cloud rancher edit

Edit the given Rancher service

```
ovhcloud cloud rancher edit <rancher_id> [flags]
```

### Options

```
  -h, --help               help for edit
      --iam-auth-enabled   Allow Rancher to use identities managed by OVHcloud IAM (Identity and Access Management) to control access
      --name string        Name of the managed Rancher service
      --plan string        Plan of the managed Rancher service (OVHCLOUD_EDITION, STANDARD)
      --version string     Version of the managed Rancher service
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud rancher](ovhcloud_cloud_rancher.md)	 - Manage Rancher services in the given cloud project

