## ovhcloud cloud reference

Fetch reference data in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for reference
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud reference container-registry](ovhcloud_cloud_reference_container-registry.md)	 - Fetch container registry reference data in the given cloud project
* [ovhcloud cloud reference database](ovhcloud_cloud_reference_database.md)	 - Fetch database reference data in the given cloud project
* [ovhcloud cloud reference list-flavors](ovhcloud_cloud_reference_list-flavors.md)	 - List available flavors in the given cloud project
* [ovhcloud cloud reference list-images](ovhcloud_cloud_reference_list-images.md)	 - List available images in the given cloud project
* [ovhcloud cloud reference loadbalancer](ovhcloud_cloud_reference_loadbalancer.md)	 - Fetch loadbalancer reference data in the given cloud project
* [ovhcloud cloud reference rancher](ovhcloud_cloud_reference_rancher.md)	 - Fetch Rancher reference data in the given cloud project

