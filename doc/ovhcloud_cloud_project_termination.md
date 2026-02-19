## ovhcloud cloud project termination

Manage project termination lifecycle

### Options

```
  -h, --help   help for termination
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

* [ovhcloud cloud project](ovhcloud_cloud_project.md)	 - Retrieve information and manage your CloudProject services
* [ovhcloud cloud project termination cancel](ovhcloud_cloud_project_termination_cancel.md)	 - Cancel a project scheduled for termination
* [ovhcloud cloud project termination confirm](ovhcloud_cloud_project_termination_confirm.md)	 - Confirm project termination with token
* [ovhcloud cloud project termination init](ovhcloud_cloud_project_termination_init.md)	 - Initiate project termination

