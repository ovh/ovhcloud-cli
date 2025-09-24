## ovhcloud cloud rancher create

Create a new Rancher service

### Synopsis

Use this command to create a managed Rancher service in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud rancher create --name MyNewRancher --plan OVHCLOUD_EDITION --version 2.11.3

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud rancher create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud rancher create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud rancher create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud rancher create --from-file ./params.json --name NameOverriden

3. Using your default text editor:

	ovhcloud cloud rancher create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud rancher create --editor --region BHS5


```
ovhcloud cloud rancher create [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for create
      --iam-auth-enabled   Allow Rancher to use identities managed by OVHcloud IAM (Identity and Access Management) to control access
      --init-file string   Create a file with example parameters
      --name string        Name of the managed Rancher service
      --plan string        Plan of the managed Rancher service (available plans can be listed using 'cloud reference rancher list-plans' command)
      --replace            Replace parameters file if it already exists
      --version string     Version of the managed Rancher service (available versions can be listed using 'cloud reference rancher list-versions' command)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud rancher](ovhcloud_cloud_rancher.md)	 - Manage Rancher services in the given cloud project

