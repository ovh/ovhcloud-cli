## ovhcloud cloud managed-analytics

Manage managed analytics services in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for managed-analytics
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
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)
* [ovhcloud cloud managed-analytics backup](ovhcloud_cloud_managed-analytics_backup.md)	 - Manage backups in a specific managed analytics service
* [ovhcloud cloud managed-analytics certificate](ovhcloud_cloud_managed-analytics_certificate.md)	 - Manage certificates in a specific managed analytics service
* [ovhcloud cloud managed-analytics create](ovhcloud_cloud_managed-analytics_create.md)	 - Create a new managed analytics service
* [ovhcloud cloud managed-analytics database](ovhcloud_cloud_managed-analytics_database.md)	 - Manage databases in a specific managed analytics service
* [ovhcloud cloud managed-analytics delete](ovhcloud_cloud_managed-analytics_delete.md)	 - Delete a specific managed analytics service
* [ovhcloud cloud managed-analytics edit](ovhcloud_cloud_managed-analytics_edit.md)	 - Edit a specific managed analytics service
* [ovhcloud cloud managed-analytics engine](ovhcloud_cloud_managed-analytics_engine.md)	 - List available analytics engines in the given cloud project
* [ovhcloud cloud managed-analytics get](ovhcloud_cloud_managed-analytics_get.md)	 - Get a specific managed analytics service
* [ovhcloud cloud managed-analytics list](ovhcloud_cloud_managed-analytics_list.md)	 - List your managed analytics services
* [ovhcloud cloud managed-analytics node-flavor](ovhcloud_cloud_managed-analytics_node-flavor.md)	 - List available analytics node flavors in the given cloud project
* [ovhcloud cloud managed-analytics pattern](ovhcloud_cloud_managed-analytics_pattern.md)	 - Manage patterns in a specific managed analytics service
* [ovhcloud cloud managed-analytics permission](ovhcloud_cloud_managed-analytics_permission.md)	 - Manage permissions in a specific managed analytics service
* [ovhcloud cloud managed-analytics plan](ovhcloud_cloud_managed-analytics_plan.md)	 - List available analytics plans in the given cloud project
* [ovhcloud cloud managed-analytics role](ovhcloud_cloud_managed-analytics_role.md)	 - Manage roles in a specific managed analytics service
* [ovhcloud cloud managed-analytics topic](ovhcloud_cloud_managed-analytics_topic.md)	 - Manage topics in a specific managed analytics service
* [ovhcloud cloud managed-analytics topic-acl](ovhcloud_cloud_managed-analytics_topic-acl.md)	 - Manage topic ACLs in a specific managed analytics service
* [ovhcloud cloud managed-analytics user](ovhcloud_cloud_managed-analytics_user.md)	 - Manage users in a specific managed analytics service

