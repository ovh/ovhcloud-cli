## ovhcloud cloud storage-swift

Manage SWIFT storage containers in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for storage-swift
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -o, --output string   Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud storage-swift edit](ovhcloud_cloud_storage-swift_edit.md)	 - Edit the given SWIFT storage container
* [ovhcloud cloud storage-swift get](ovhcloud_cloud_storage-swift_get.md)	 - Get a specific SWIFT storage container
* [ovhcloud cloud storage-swift list](ovhcloud_cloud_storage-swift_list.md)	 - List SWIFT storage containers

