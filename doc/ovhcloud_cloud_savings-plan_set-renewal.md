## ovhcloud cloud savings-plan set-renewal

Set the action at the end of the savings plan period

### Synopsis

Set the action to be performed when the savings plan reaches the end of its period.

Available actions:
- REACTIVATE: Automatically renew the savings plan for another period
- TERMINATE: Terminate the savings plan at the end of the period

```
ovhcloud cloud savings-plan set-renewal <savings_plan_id> [flags]
```

### Options

```
      --action string   Action at period end: REACTIVATE or TERMINATE (required)
  -h, --help            help for set-renewal
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

