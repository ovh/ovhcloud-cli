## ovhcloud cloud savings-plan simulate

Simulate a savings plan subscription

### Synopsis

Simulate subscribing to an OVHcloud Savings Plan without actually subscribing.

This is useful to preview what the savings plan would look like before committing.
You can use either --offer-id or --flavor with --deployment-type.

```
ovhcloud cloud savings-plan simulate [flags]
```

### Options

```
      --deployment-type string   Deployment type: 1AZ or 3AZ (default: 1AZ) (default "1AZ")
      --display-name string      Custom display name (required)
      --flavor string            Savings plan flavor (e.g., b3-8, rancher, c3-16)
  -h, --help                     help for simulate
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

