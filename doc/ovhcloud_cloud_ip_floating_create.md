## ovhcloud cloud ip floating create

Create a new floating IP

### Synopsis

Use this command to create a floating IP in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud ip floating create --region GRA11 --description "My floating IP"

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud ip floating create --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud ip floating create --from-file ./params.json

3. Using your default text editor:

	ovhcloud cloud ip floating create --editor


```
ovhcloud cloud ip floating create [flags]
```

### Options

```
      --availability-zone string   Availability zone within the region
      --description string         Description of the floating IP
      --editor                     Use a text editor to define parameters
      --from-file string           File containing parameters
  -h, --help                       help for create
      --init-file string           Create a file with example parameters
      --region string              Region where the floating IP will be created
      --replace                    Replace parameters file if it already exists
      --wait                       Wait for floating IP creation to be done before exiting
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

* [ovhcloud cloud ip floating](ovhcloud_cloud_ip_floating.md)	 - Manage floating public IPs in the given cloud project

