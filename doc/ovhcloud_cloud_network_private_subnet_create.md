## ovhcloud cloud network private subnet create

Create a subnet in the given private network

### Synopsis

Use this command to create a new subnet in a private network.
There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud network private subnet create <network_id> --network 192.168.1.0/24 --start 192.168.1.12 --end 192.168.1.24 --region GRA9

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network private subnet create <network_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network private subnet create <network_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network private subnet create <network_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network private subnet create <network_id> --from-file ./params.json --region BHS5

3. Using your default text editor:

	ovhcloud cloud network private subnet create <network_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud network private subnet create <network_id> --editor --region DE1


```
ovhcloud cloud network private subnet create <network_id> [flags]
```

### Options

```
      --dhcp               Enable DHCP for the subnet
      --editor             Use a text editor to define parameters
      --end string         Last IP for this region (eg: 192.168.1.24)
      --from-file string   File containing parameters
  -h, --help               help for create
      --init-file string   Create a file with example parameters
      --network string     Global network CIDR (eg: 192.168.1.0/24)
      --no-gateway         Use this flag if you don't want to set a default gateway IP
      --region string      Region for the subnet
      --replace            Replace parameters file if it already exists
      --start string       First IP for this region (eg: 192.168.1.12)
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

* [ovhcloud cloud network private subnet](ovhcloud_cloud_network_private_subnet.md)	 - Manage subnets in a specific private network

