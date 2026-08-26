## ovhcloud cloud network security-group create

Create a new security group

### Synopsis

Use this command to create a security group in the given public cloud project.

Rules are nested objects: to define them, use a configuration file or your text
editor. There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud network security-group create --name my-sg --region GRA11

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud network security-group create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud network security-group create --from-file ./params.json

  You can also pipe the content of the parameters file:

	cat ./params.json | ovhcloud cloud network security-group create

  In both cases, you can override values using command line flags, for example:

	ovhcloud cloud network security-group create --from-file ./params.json --name my-sg

3. Using your default text editor:

	ovhcloud cloud network security-group create --editor


```
ovhcloud cloud network security-group create [flags]
```

### Options

```
      --availability-zone string   Availability zone within the region
      --description string         Description of the security group
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --name string                Name of the security group
      --region string              Region where the security group will be created
      --replace                    Replace parameters file if it already exists
      --wait                       Wait for security group creation to be done before exiting
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

* [ovhcloud cloud network security-group](ovhcloud_cloud_network_security-group.md)	 - Manage security groups in the given cloud project

