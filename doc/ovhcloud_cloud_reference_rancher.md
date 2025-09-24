## ovhcloud cloud reference rancher

Fetch Rancher reference data in the given cloud project

### Options

```
  -h, --help   help for rancher
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

* [ovhcloud cloud reference](ovhcloud_cloud_reference.md)	 - Fetch reference data in the given cloud project
* [ovhcloud cloud reference rancher list-plans](ovhcloud_cloud_reference_rancher_list-plans.md)	 - List available Rancher plans in the given cloud project
* [ovhcloud cloud reference rancher list-versions](ovhcloud_cloud_reference_rancher_list-versions.md)	 - List available Rancher versions in the given cloud project

