## ovhcloud cloud savings-plan

Manage savings plans for your cloud project

### Synopsis

Manage OVHcloud Savings Plans for your Public Cloud project.

Savings Plans allow you to commit to a consistent amount of usage (measured in $/hour) 
for a 1-month term, in exchange for discounted pricing on your cloud resources.

Available flavors include:
- Rancher: rancher, rancher_standard, rancher_ovhcloud_edition
- General purpose instances: b3-8, b3-16, b3-32, b3-64, b3-128, b3-256
- Compute optimized instances: c3-4, c3-8, c3-16, c3-32, c3-64, c3-128
- Memory optimized instances: r3-16, r3-32, r3-64, r3-128, r3-256, r3-512

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for savings-plan
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
* [ovhcloud cloud savings-plan edit](ovhcloud_cloud_savings-plan_edit.md)	 - Edit a savings plan's display name
* [ovhcloud cloud savings-plan get](ovhcloud_cloud_savings-plan_get.md)	 - Get details of a specific savings plan
* [ovhcloud cloud savings-plan list](ovhcloud_cloud_savings-plan_list.md)	 - List subscribed savings plans
* [ovhcloud cloud savings-plan list-offers](ovhcloud_cloud_savings-plan_list-offers.md)	 - List available savings plan offers to subscribe to
* [ovhcloud cloud savings-plan list-periods](ovhcloud_cloud_savings-plan_list-periods.md)	 - List the period history of a savings plan
* [ovhcloud cloud savings-plan resize](ovhcloud_cloud_savings-plan_resize.md)	 - Change the size of a savings plan
* [ovhcloud cloud savings-plan set-renewal](ovhcloud_cloud_savings-plan_set-renewal.md)	 - Set the action at the end of the savings plan period
* [ovhcloud cloud savings-plan simulate](ovhcloud_cloud_savings-plan_simulate.md)	 - Simulate a savings plan subscription
* [ovhcloud cloud savings-plan subscribe](ovhcloud_cloud_savings-plan_subscribe.md)	 - Subscribe to a new savings plan
* [ovhcloud cloud savings-plan terminate](ovhcloud_cloud_savings-plan_terminate.md)	 - Terminate/unsubscribe from a savings plan

