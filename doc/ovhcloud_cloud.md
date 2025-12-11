## ovhcloud cloud

Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)

### Options

```
  -h, --help   help for cloud
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud cloud container-registry](ovhcloud_cloud_container-registry.md)	 - Manage container registries in the given cloud project
* [ovhcloud cloud database-service](ovhcloud_cloud_database-service.md)	 - Manage database services in the given cloud project
* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project
* [ovhcloud cloud kube](ovhcloud_cloud_kube.md)	 - Manage Kubernetes clusters in the given cloud project
* [ovhcloud cloud network](ovhcloud_cloud_network.md)	 - Manage networks in the given cloud project
* [ovhcloud cloud operation](ovhcloud_cloud_operation.md)	 - List and get operations in the given cloud project
* [ovhcloud cloud project](ovhcloud_cloud_project.md)	 - Retrieve information and manage your CloudProject services
* [ovhcloud cloud quota](ovhcloud_cloud_quota.md)	 - Check quotas in the given cloud project
* [ovhcloud cloud rancher](ovhcloud_cloud_rancher.md)	 - Manage Rancher services in the given cloud project
* [ovhcloud cloud reference](ovhcloud_cloud_reference.md)	 - Fetch reference data in the given cloud project
* [ovhcloud cloud region](ovhcloud_cloud_region.md)	 - Check regions in the given cloud project
* [ovhcloud cloud ssh-key](ovhcloud_cloud_ssh-key.md)	 - Manage SSH keys in the given cloud project
* [ovhcloud cloud storage-block](ovhcloud_cloud_storage-block.md)	 - Manage block storage volumes in the given cloud project
* [ovhcloud cloud storage-s3](ovhcloud_cloud_storage-s3.md)	 - Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)
* [ovhcloud cloud storage-swift](ovhcloud_cloud_storage-swift.md)	 - Manage SWIFT storage containers in the given cloud project
* [ovhcloud cloud user](ovhcloud_cloud_user.md)	 - Manage users in the given cloud project

