## ovhcloud cloud private-endpoint create

Create a new private endpoint

### Synopsis

Use this command to create a new private endpoint in the given public cloud project.

Use the "service-availability list" command to discover the services and tiers
that can be exposed. There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud private-endpoint create --name MyEndpoint --region GRA11 --service-name MyService --tier S --subnet-id <subnet_id>

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud private-endpoint create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud private-endpoint create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud private-endpoint create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud private-endpoint create --from-file ./params.json --name MyEndpoint

3. Using your default text editor:

	ovhcloud cloud private-endpoint create --editor


```
ovhcloud cloud private-endpoint create [flags]
```

### Options

```
      --description string    Description of the private endpoint
      --editor                Use a text editor to define parameters
      --from-file string      File containing parameters
  -h, --help                  help for create
      --init-file string      Create a file with example parameters
      --name string           Name of the private endpoint
      --network-id string     Network to allocate the endpoint port in
      --public-dns-enabled    Register the endpoint hostname in the public DNS zone
      --region string         Region where the private endpoint will be created
      --replace               Replace parameters file if it already exists
      --service-name string   Name of the service to expose
      --subnet-id string      Subnet to allocate the endpoint port in
      --tier string           Tier of the service (available tiers can be listed using 'ovhcloud cloud private-endpoint service-availability list' command)
      --vip-address string    Private IP address to assign to the endpoint inside the subnet
      --wait                  Wait for private endpoint creation to be done before exiting
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
```

### SEE ALSO

* [ovhcloud cloud private-endpoint](ovhcloud_cloud_private-endpoint.md)	 - Manage private endpoints in the given cloud project

