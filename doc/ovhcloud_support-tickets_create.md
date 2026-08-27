## ovhcloud support-tickets create

Create a new support ticket

### Synopsis

Use this command to create a new support ticket.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud support-tickets create --subject 'My issue' --body 'Detailed description of the issue' --product publiccloud --category assistance

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud support-tickets create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud support-tickets create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud support-tickets create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud support-tickets create --from-file ./params.json --subject 'Overridden subject'

3. Using your default text editor:

	ovhcloud support-tickets create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud support-tickets create --editor --subject 'Overridden subject'


```
ovhcloud support-tickets create [flags]
```

### Options

```
      --body string           Ticket message body
      --category string       Ticket category (assistance, billing, incident)
      --editor                Use a text editor to define parameters
      --from-file string      File containing parameters
  -h, --help                  help for create
      --impact string         Ticket impact - Business/Enterprise support only (low, medium, high)
      --init-file string      Create a file with example parameters
      --product string        Ticket product (adsl, cdn, dedicated, dedicated-billing, dedicated-other, dedicatedcloud, domain, exchange, fax, hosting, housing, iaas, mail, network, publiccloud, sms, ssl, storage, telecom-billing, telecom-other, vac, voip, vps, web-billing, web-other)
      --replace               Replace parameters file if it already exists
      --service-name string   Service name (resource identifier) the ticket is about
      --subcategory string    Ticket subcategory (alerts, autorenew, bill, down, inProgress, new, other, perfs, start, usage)
      --subject string        Ticket subject (short summary)
      --urgency string        Ticket urgency - Business/Enterprise support only (low, medium, high)
      --watchers strings      Comma-separated list of e-mail addresses to notify on ticket updates (max. 10)
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
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud support-tickets](ovhcloud_support-tickets.md)	 - Retrieve information and manage your support tickets

