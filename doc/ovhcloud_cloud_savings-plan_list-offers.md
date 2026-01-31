## ovhcloud cloud savings-plan list-offers

List available savings plan offers to subscribe to

### Synopsis

List available savings plan offers that can be subscribed to.

Use --product-code to filter by flavor (e.g., 'b3-8', 'rancher', 'c3-16').
Use --deployment-type to filter by availability zone configuration (1AZ or 3AZ).

Note: Rancher flavors only support 1AZ deployment.

```
ovhcloud cloud savings-plan list-offers [flags]
```

### Options

```
      --deployment-type string   Deployment type: 1AZ or 3AZ (default: 1AZ) (default "1AZ")
      --filter stringArray       Filter results by any property using https://github.com/PaesslerAG/gval syntax
                                 Examples:
                                   --filter 'state="running"'
                                   --filter 'name=~"^my.*"'
                                   --filter 'nested.property.subproperty>10'
                                   --filter 'startDate>="2023-12-01"'
                                   --filter 'name=~"something" && nbField>10'
  -h, --help                     help for list-offers
      --product-code string      Filter offers by product code (e.g., 'b3-8', 'rancher')
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud savings-plan](ovhcloud_cloud_savings-plan.md)	 - Manage savings plans for your cloud project

