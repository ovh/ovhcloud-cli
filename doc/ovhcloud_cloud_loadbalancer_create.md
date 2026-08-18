## ovhcloud cloud loadbalancer create

Create a loadbalancer in the given cloud project

### Synopsis

Use this command to create a loadbalancer.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer create <region> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer create <region> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer create <region> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer create <region> --name my-lb --size small


```
ovhcloud cloud loadbalancer create <region> [flags]
```

### Options

```
      --editor               Use a text editor to define parameters
      --floating-ip string   Floating IP ID to associate to the loadbalancer
      --from-file string     File containing parameters
      --gateway string       Gateway ID to associate to the loadbalancer
  -h, --help                 help for create
      --init-file string     Create a file with example parameters
      --name string          Name of the loadbalancer
      --network-id string    Network ID
      --replace              Replace parameters file if it already exists
      --size string          Size of the loadbalancer (e.g. small, medium, large) or flavor UUID
      --subnet-id string     Subnet ID
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
      --raw                    Output the extracted value without JSON quoting (use with -o '<field>'), useful for scripting
                               Example:
                                 --output 'id' --raw   (prints the id without surrounding quotes)
```

### SEE ALSO

* [ovhcloud cloud loadbalancer](ovhcloud_cloud_loadbalancer.md)	 - Manage loadbalancers in the given cloud project

