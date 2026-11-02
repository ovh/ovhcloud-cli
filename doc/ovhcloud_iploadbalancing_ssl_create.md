## ovhcloud iploadbalancing ssl create

Add a new custom SSL certificate on the given IpLoadbalancing

### Synopsis

Use this command to add a new custom SSL certificate on the given IpLoadbalancing.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud iploadbalancing ssl create <service_name> --certificate "$(cat cert.pem)" --key "$(cat key.pem)"

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud iploadbalancing ssl create <service_name> --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud iploadbalancing ssl create <service_name> --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud iploadbalancing ssl create <service_name>

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud iploadbalancing ssl create <service_name> --from-file ./params.json --display-name nameoverriden

3. Using your default text editor:

	ovhcloud iploadbalancing ssl create <service_name> --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud iploadbalancing ssl create <service_name> --editor --display-name nameoverriden


```
ovhcloud iploadbalancing ssl create <service_name> [flags]
```

### Options

```
      --certificate string    SSL certificate (PEM)
      --chain string          SSL certificate chain (PEM)
      --display-name string   Human readable name for your SSL certificate
      --editor                Use a text editor to define parameters
      --from-file string      File containing parameters
  -h, --help                  help for create
      --init-file string      Create a file with example parameters
      --key string            SSL private key (PEM)
      --replace               Replace parameters file if it already exists
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
```

### SEE ALSO

* [ovhcloud iploadbalancing ssl](ovhcloud_iploadbalancing_ssl.md)	 - Manage SSL certificates of your IpLoadbalancing services

