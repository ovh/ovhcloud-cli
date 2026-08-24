## ovhcloud baremetal

Retrieve information and manage your Bare Metal services

### Options

```
  -h, --help   help for baremetal
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

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud baremetal backup](ovhcloud_baremetal_backup.md)	 - Manage the backup spaces of a dedicated server
* [ovhcloud baremetal backup-agent](ovhcloud_baremetal_backup-agent.md)	 - Manage the Veeam backup agent protecting this server
* [ovhcloud baremetal boot](ovhcloud_baremetal_boot.md)	 - Manage boot options for the given baremetal
* [ovhcloud baremetal catalog](ovhcloud_baremetal_catalog.md)	 - List orderable servers, their availability and their price
* [ovhcloud baremetal confirm-termination](ovhcloud_baremetal_confirm-termination.md)	 - Confirm the termination of the given baremetal
* [ovhcloud baremetal cost](ovhcloud_baremetal_cost.md)	 - Show what a server costs and when it renews
* [ovhcloud baremetal doctor](ovhcloud_baremetal_doctor.md)	 - Report what is wrong with a server, or with every server
* [ovhcloud baremetal edit](ovhcloud_baremetal_edit.md)	 - Update the given baremetal
* [ovhcloud baremetal get](ovhcloud_baremetal_get.md)	 - Retrieve information of a specific baremetal
* [ovhcloud baremetal install-status](ovhcloud_baremetal_install-status.md)	 - Show how far the running installation of this baremetal has got
* [ovhcloud baremetal ipmi](ovhcloud_baremetal_ipmi.md)	 - Manage IPMI on your baremetal
* [ovhcloud baremetal list](ovhcloud_baremetal_list.md)	 - List your Baremetal services
* [ovhcloud baremetal list-compatible-os](ovhcloud_baremetal_list-compatible-os.md)	 - Retrieve OSes that can be installed on this baremetal
* [ovhcloud baremetal list-interventions](ovhcloud_baremetal_list-interventions.md)	 - List past and planned interventions for the given baremetal
* [ovhcloud baremetal list-ips](ovhcloud_baremetal_list-ips.md)	 - List all IPs that are routed to the given baremetal
* [ovhcloud baremetal list-partition-schemes](ovhcloud_baremetal_list-partition-schemes.md)	 - List the partition schemes an OS template allows on this baremetal
* [ovhcloud baremetal list-secrets](ovhcloud_baremetal_list-secrets.md)	 - Retrieve secrets to connect to the server
* [ovhcloud baremetal list-tasks](ovhcloud_baremetal_list-tasks.md)	 - Retrieve tasks of the given baremetal
* [ovhcloud baremetal logs](ovhcloud_baremetal_logs.md)	 - Read the logs of a dedicated server, and send them to a stream
* [ovhcloud baremetal power](ovhcloud_baremetal_power.md)	 - Power the given baremetal off and on
* [ovhcloud baremetal raid-profile](ovhcloud_baremetal_raid-profile.md)	 - Show the hardware RAID controllers of this baremetal, if it has any
* [ovhcloud baremetal reboot](ovhcloud_baremetal_reboot.md)	 - Reboot the given baremetal
* [ovhcloud baremetal reboot-rescue](ovhcloud_baremetal_reboot-rescue.md)	 - Reboot the given baremetal in rescue mode
* [ovhcloud baremetal reinstall](ovhcloud_baremetal_reinstall.md)	 - Reinstall the given baremetal
* [ovhcloud baremetal service-info](ovhcloud_baremetal_service-info.md)	 - Manage service information of the given baremetal
* [ovhcloud baremetal terminate](ovhcloud_baremetal_terminate.md)	 - Ask for the termination of the given baremetal
* [ovhcloud baremetal ticket](ovhcloud_baremetal_ticket.md)	 - Open a support ticket about a server, with its state already in it
* [ovhcloud baremetal traffic](ovhcloud_baremetal_traffic.md)	 - Show the traffic graphs of this baremetal's network controllers
* [ovhcloud baremetal vni](ovhcloud_baremetal_vni.md)	 - Manage Virtual Network Interfaces of the given baremetal
* [ovhcloud baremetal vrack](ovhcloud_baremetal_vrack.md)	 - Attach the given baremetal to a vRack, or detach it

