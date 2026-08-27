## ovhcloud cloud network gateway create

Create a gateway in the given cloud project

### Synopsis

Use this command to create a new gateway in the given public cloud project.

Subnets are nested objects: to attach them, use the repeatable --subnet flag, a
configuration file or your text editor. There are three ways to define the
creation parameters:

1. Using only CLI flags:

	ovhcloud cloud network gateway create --name MyGateway --region GRA11 --external-gateway-enabled --external-gateway-model S

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network gateway create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network gateway create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud network gateway create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud network gateway create --from-file ./params.json --name MyGateway

3. Using your default text editor:

	ovhcloud cloud network gateway create --editor


```
ovhcloud cloud network gateway create [flags]
```

### Options

```
      --availability-zone string        Availability zone within the region
      --description string              Description of the gateway
      --editor                          Use a text editor to define parameters
      --external-gateway-enabled        Whether the external gateway is enabled
      --external-gateway-model string   External gateway sizing model (S, M, L, XL, 2XL, 3XL)
      --from-file string                File containing parameters
  -h, --help                            help for create
      --init-file string                Create a file with example parameters
      --name string                     Name of the gateway
      --region string                   Region where the gateway will be created
      --replace                         Replace parameters file if it already exists
      --subnet strings                  ID of a subnet to attach to the gateway (repeatable)
      --wait                            Wait for gateway creation to be done before exiting
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

* [ovhcloud cloud network gateway](ovhcloud_cloud_network_gateway.md)	 - Manage gateways in the given cloud project

