## ovhcloud cloud network private vrack create

Create a private network in the given cloud project

### Synopsis

Use this command to create a private network.
There are three ways to define the parameters:

1. Using only CLI flags:

	ovhcloud cloud network private vrack create <region> --name MyNetwork

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network private vrack create <region> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network private vrack create <region> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network private vrack create <region>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network private vrack create <region> --from-file ./params.json --name MyNetwork

3. Using your default text editor:

	ovhcloud cloud network private vrack create <region> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud network private vrack create <region> --editor --name MyNetwork


```
ovhcloud cloud network private vrack create <region> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for create
      --init-file string   Create a file with example parameters
      --name string        Name of the private network
      --replace            Replace parameters file if it already exists
      --vlan-id int        VLAN ID for the private network
      --wait               Wait for network creation to be done before exiting
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
      --profile string         Use a specific profile from the configuration file
      --region string          Filter by region or specify the region of the network
```

### SEE ALSO

* [ovhcloud cloud network private vrack](ovhcloud_cloud_network_private_vrack.md)	 - Manage vRack-based private networks in the given cloud project

