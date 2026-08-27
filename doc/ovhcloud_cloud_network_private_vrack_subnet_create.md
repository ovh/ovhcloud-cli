## ovhcloud cloud network private vrack subnet create

Create a subnet in the given private network

### Synopsis

Use this command to create a new subnet in a private network.
There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud network private vrack subnet create <network_id> --name MySubnet --cidr 192.168.1.0/24

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network private vrack subnet create <network_id> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network private vrack subnet create <network_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network private vrack subnet create <network_id>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network private vrack subnet create <network_id> --from-file ./params.json --name MySubnet

3. Using your default text editor:

	ovhcloud cloud network private vrack subnet create <network_id> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud network private vrack subnet create <network_id> --editor --name MySubnet


```
ovhcloud cloud network private vrack subnet create <network_id> [flags]
```

### Options

```
      --allocation-pools strings   Allocation pools for the subnet in format start:end
      --availability-zone string   Availability zone within the region
      --cidr string                CIDR of the subnet (eg: 192.168.1.0/24)
      --description string         Description of the subnet
      --dhcp-enabled               Enable DHCP for the subnet
      --dns-nameservers strings    DNS nameservers for the subnet
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
      --gateway-ip string          Gateway IP address for the subnet
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Name of the subnet
      --region string              Region of the subnet (defaults to the parent network region)
      --replace                    Replace parameters file if it already exists
      --wait                       Wait for subnet creation to be done before exiting
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud network private vrack subnet](ovhcloud_cloud_network_private_vrack_subnet.md)	 - Manage subnets in a specific private network

