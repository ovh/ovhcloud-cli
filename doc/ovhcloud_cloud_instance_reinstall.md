## ovhcloud cloud instance reinstall

Reinstall the given instance

### Synopsis

Use this command to reinstall the given instance.

There are three ways to define the installation parameters:
(the following examples assume that you have already configured your default cloud project using "ovhcloud config set default_cloud_project <project_id>")

1. Using only CLI flags:

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --image <image_id>

2. Using the interactive image selector:

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --image-selector

3. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud instance reinstall --init-file ./install.json

  You will be able to choose from several installation examples. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct installation parameters, run:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

	cat ./install.json | ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --from-file ./install.json --image <image_id>

4. Using your default text editor:

  ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --editor

  You will be able to choose from several installation examples. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reinstallation will be run.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --editor --image <image_id>


```
ovhcloud cloud instance reinstall <instance_id> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for reinstall
      --image string       Image to use for reinstallation
      --image-selector     Use the interactive image selector to define installation parameters
      --init-file string   Create a file with example parameters
      --replace            Replace parameters file if it already exists
      --wait               Wait for reinstall to be done before exiting
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

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project

