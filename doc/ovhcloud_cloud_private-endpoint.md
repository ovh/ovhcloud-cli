## ovhcloud cloud private-endpoint

Manage private endpoints in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for private-endpoint
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud private-endpoint create](ovhcloud_cloud_private-endpoint_create.md)	 - Create a new private endpoint
* [ovhcloud cloud private-endpoint delete](ovhcloud_cloud_private-endpoint_delete.md)	 - Delete a specific private endpoint
* [ovhcloud cloud private-endpoint edit](ovhcloud_cloud_private-endpoint_edit.md)	 - Edit the given private endpoint (only name and description are mutable)
* [ovhcloud cloud private-endpoint get](ovhcloud_cloud_private-endpoint_get.md)	 - Get a specific private endpoint
* [ovhcloud cloud private-endpoint list](ovhcloud_cloud_private-endpoint_list.md)	 - List private endpoints
* [ovhcloud cloud private-endpoint service-availability](ovhcloud_cloud_private-endpoint_service-availability.md)	 - Consult the services that can be exposed through a private endpoint

