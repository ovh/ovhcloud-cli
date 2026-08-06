## ovhcloud ip firewall rule create

Create a new firewall rule

### Synopsis

Use this command to create a new firewall rule.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud ip firewall rule create <ip_block> <ip> --action permit --protocol tcp --sequence 0 --destination-port 443

2. Using a configuration file:

  First generate an example parameters file:

	ovhcloud ip firewall rule create <ip_block> <ip> --init-file ./rule.json

  After editing the file, run:

	ovhcloud ip firewall rule create <ip_block> <ip> --from-file ./rule.json

  You can also pipe the content:

	cat ./rule.json | ovhcloud ip firewall rule create <ip_block> <ip>

3. Using your default text editor:

	ovhcloud ip firewall rule create <ip_block> <ip> --editor


```
ovhcloud ip firewall rule create <ip_block> <ip> [flags]
```

### Options

```
      --action string          Action: deny or permit (required)
      --destination-port int   Destination port (TCP/UDP only)
      --editor                 Use a text editor to define parameters
      --from-file string       File containing parameters
  -h, --help                   help for create
      --init-file string       Create a file with example parameters
      --protocol string        Protocol: ah, esp, gre, icmp, ipv4, tcp, udp (required)
      --replace                Replace parameters file if it already exists
      --sequence int           Rule priority 0-19 (required) (default -1)
      --source string          Source IP/CIDR (defaults to any)
      --source-port int        Source port (TCP/UDP only)
      --tcp-fragments          TCP fragments option
      --tcp-option string      TCP option: established or syn (TCP only)
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
  -y, --yes              Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud ip firewall rule](ovhcloud_ip_firewall_rule.md)	 - Manage firewall rules

