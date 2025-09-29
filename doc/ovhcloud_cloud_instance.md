## ovhcloud cloud instance

Manage instances in the given cloud project

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for instance
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
* [ovhcloud cloud instance activate-monthly-billing](ovhcloud_cloud_instance_activate-monthly-billing.md)	 - Activate monthly billing for the given instance
* [ovhcloud cloud instance create](ovhcloud_cloud_instance_create.md)	 - Create a new instance
* [ovhcloud cloud instance delete](ovhcloud_cloud_instance_delete.md)	 - Delete the given instance
* [ovhcloud cloud instance exit-rescue](ovhcloud_cloud_instance_exit-rescue.md)	 - Exit the given instance from rescue mode
* [ovhcloud cloud instance get](ovhcloud_cloud_instance_get.md)	 - Get a specific instance
* [ovhcloud cloud instance interface](ovhcloud_cloud_instance_interface.md)	 - Manage interfaces of the given instance
* [ovhcloud cloud instance list](ovhcloud_cloud_instance_list.md)	 - List your instances
* [ovhcloud cloud instance reboot](ovhcloud_cloud_instance_reboot.md)	 - Reboot the given instance
* [ovhcloud cloud instance reboot-rescue](ovhcloud_cloud_instance_reboot-rescue.md)	 - Reboot the given instance in rescue mode
* [ovhcloud cloud instance reinstall](ovhcloud_cloud_instance_reinstall.md)	 - Reinstall the given instance
* [ovhcloud cloud instance resume](ovhcloud_cloud_instance_resume.md)	 - Resume the given suspended instance
* [ovhcloud cloud instance set-flavor](ovhcloud_cloud_instance_set-flavor.md)	 - Migrate the given instance to the specified flavor
* [ovhcloud cloud instance set-name](ovhcloud_cloud_instance_set-name.md)	 - Set the name of the given instance
* [ovhcloud cloud instance shelve](ovhcloud_cloud_instance_shelve.md)	 - Shelve the given instance
* [ovhcloud cloud instance snapshot](ovhcloud_cloud_instance_snapshot.md)	 - Manage snapshots of the given instance
* [ovhcloud cloud instance start](ovhcloud_cloud_instance_start.md)	 - Start the given instance
* [ovhcloud cloud instance stop](ovhcloud_cloud_instance_stop.md)	 - Stop the given instance
* [ovhcloud cloud instance unshelve](ovhcloud_cloud_instance_unshelve.md)	 - Unshelve the given instance

