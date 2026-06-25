## ovhcloud cloud managed-database

Manage managed database services in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for managed-database
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
* [ovhcloud cloud managed-database backup](ovhcloud_cloud_managed-database_backup.md)	 - Manage backups in a specific managed database service
* [ovhcloud cloud managed-database certificate](ovhcloud_cloud_managed-database_certificate.md)	 - Manage certificates in a specific managed database service
* [ovhcloud cloud managed-database create](ovhcloud_cloud_managed-database_create.md)	 - Create a new managed database service
* [ovhcloud cloud managed-database database](ovhcloud_cloud_managed-database_database.md)	 - Manage databases in a specific managed database service
* [ovhcloud cloud managed-database delete](ovhcloud_cloud_managed-database_delete.md)	 - Delete a specific managed database service
* [ovhcloud cloud managed-database edit](ovhcloud_cloud_managed-database_edit.md)	 - Edit a specific managed database service
* [ovhcloud cloud managed-database get](ovhcloud_cloud_managed-database_get.md)	 - Get a specific managed database service
* [ovhcloud cloud managed-database list](ovhcloud_cloud_managed-database_list.md)	 - List your managed database services
* [ovhcloud cloud managed-database list-engines](ovhcloud_cloud_managed-database_list-engines.md)	 - List available database engines in the given cloud project
* [ovhcloud cloud managed-database list-node-flavors](ovhcloud_cloud_managed-database_list-node-flavors.md)	 - List available database node flavors in the given cloud project
* [ovhcloud cloud managed-database list-plans](ovhcloud_cloud_managed-database_list-plans.md)	 - List available database plans in the given cloud project
* [ovhcloud cloud managed-database role](ovhcloud_cloud_managed-database_role.md)	 - Manage roles in a specific managed database service
* [ovhcloud cloud managed-database user](ovhcloud_cloud_managed-database_user.md)	 - Manage users in a specific managed database service

