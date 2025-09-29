## ovhcloud baremetal

Retrieve information and manage your Bare Metal services

### Options

```
  -h, --help   help for baremetal
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
* [ovhcloud baremetal boot](ovhcloud_baremetal_boot.md)	 - Manage boot options for the given baremetal
* [ovhcloud baremetal edit](ovhcloud_baremetal_edit.md)	 - Update the given baremetal
* [ovhcloud baremetal get](ovhcloud_baremetal_get.md)	 - Retrieve information of a specific baremetal
* [ovhcloud baremetal ipmi](ovhcloud_baremetal_ipmi.md)	 - Manage IPMI on your baremetal
* [ovhcloud baremetal list](ovhcloud_baremetal_list.md)	 - List your Baremetal services
* [ovhcloud baremetal list-compatible-os](ovhcloud_baremetal_list-compatible-os.md)	 - Retrieve OSes that can be installed on this baremetal
* [ovhcloud baremetal list-interventions](ovhcloud_baremetal_list-interventions.md)	 - List past and planned interventions for the given baremetal
* [ovhcloud baremetal list-ips](ovhcloud_baremetal_list-ips.md)	 - List all IPs that are routed to the given baremetal
* [ovhcloud baremetal list-secrets](ovhcloud_baremetal_list-secrets.md)	 - Retrieve secrets to connect to the server
* [ovhcloud baremetal list-tasks](ovhcloud_baremetal_list-tasks.md)	 - Retrieve tasks of the given baremetal
* [ovhcloud baremetal reboot](ovhcloud_baremetal_reboot.md)	 - Reboot the given baremetal
* [ovhcloud baremetal reboot-rescue](ovhcloud_baremetal_reboot-rescue.md)	 - Reboot the given baremetal in rescue mode
* [ovhcloud baremetal reinstall](ovhcloud_baremetal_reinstall.md)	 - Reinstall the given baremetal
* [ovhcloud baremetal vni](ovhcloud_baremetal_vni.md)	 - Manage Virtual Network Interfaces of the given baremetal

