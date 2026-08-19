## ovhcloud cloud savings-plan terminate

Terminate/unsubscribe from a savings plan

### Synopsis

Terminate an existing savings plan subscription.

By default, the savings plan will be terminated at the end of its current period.
You can specify a termination date using the --termination-date flag.

```
ovhcloud cloud savings-plan terminate <savings_plan_id> [flags]
```

### Options

```
  -h, --help                      help for terminate
      --termination-date string   Termination date (YYYY-MM-DD format, optional)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud savings-plan](ovhcloud_cloud_savings-plan.md)	 - Manage savings plans for your cloud project

