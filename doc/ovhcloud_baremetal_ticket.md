## ovhcloud baremetal ticket

Open a support ticket about a server, with its state already in it

### Synopsis

Create a support ticket that names the server and carries its current state:
commercial range, datacenter, active boot, monitoring, running tasks, planned
maintenance, and whatever "ovhcloud baremetal doctor" reports about it.

The description is yours - the collected block says what the machine is, not
what is wrong from your side. Use --no-context to send the description alone.

A ticket is read by a person at OVHcloud, so the command confirms before
sending. --dry-run prints the whole message instead of sending it.

```
ovhcloud baremetal ticket <service_name> [flags]
```

### Options

```
      --body string          Describe the problem in your own words
      --category string      Ticket category
      --dry-run              Print the whole ticket that would be created without creating it
  -h, --help                 help for ticket
      --impact string        Ticket impact - Business/Enterprise support only
      --no-context           Send the description without the collected machine state
      --product string       Product the ticket is filed under (default "dedicated")
      --subcategory string   Ticket subcategory
      --subject string       Ticket subject (short summary)
      --urgency string       Ticket urgency - Business/Enterprise support only
      --watchers strings     E-mail addresses to notify on updates (max. 10)
  -y, --yes                  Skip the confirmation prompt (required for unattended runs)
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

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your Bare Metal services

