## ovhcloud cloud savings-plan subscribe

Subscribe to a new savings plan

### Synopsis

Subscribe to a new OVHcloud Savings Plan.

You can subscribe in two ways:

1. Using an offer ID directly:
   ovhcloud cloud savings-plan subscribe --offer-id <offer_id> --display-name "My Plan" --size 2

2. Using flavor and deployment type (the CLI will find the matching offer):
   ovhcloud cloud savings-plan subscribe --flavor b3-8 --deployment-type 1AZ --display-name "My Plan" --size 2

Available flavors:
- Rancher: rancher, rancher_standard, rancher_ovhcloud_edition (1AZ only)
- General purpose: b3-8, b3-16, b3-32, b3-64, b3-128, b3-256
- Compute optimized: c3-4, c3-8, c3-16, c3-32, c3-64, c3-128
- Memory optimized: r3-16, r3-32, r3-64, r3-128, r3-256, r3-512

Deployment types:
- 1AZ: Single availability zone (default)
- 3AZ: Three availability zones (not available for Rancher)

```
ovhcloud cloud savings-plan subscribe [flags]
```

### Options

```
      --deployment-type string   Deployment type: 1AZ or 3AZ (default: 1AZ) (default "1AZ")
      --display-name string      Custom display name (required)
      --flavor string            Savings plan flavor (e.g., b3-8, rancher, c3-16)
  -h, --help                     help for subscribe
      --offer-id string          Offer ID from list-offers (alternative to --flavor)
      --size int                 Size of the savings plan (required)
      --start-date string        Start date (YYYY-MM-DD format, defaults to today)
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

