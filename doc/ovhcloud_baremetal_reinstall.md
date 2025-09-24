## ovhcloud baremetal reinstall

Reinstall the given baremetal

### Synopsis

Use this command to reinstall the given dedicated server.
There are three ways to define the installation parameters:

1. Using only CLI flags:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --os byolinux_64 --language fr-fr --image-url https://...

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud baremetal reinstall --init-file ./install.json

  You will be able to choose from several installation examples. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct installation parameters, run:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json

  Note that you can also pipe the content of the file to reinstall, like the following:

	cat ./install.json | ovhcloud baremetal reinstall ns1234.ip-11.22.33.net

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --from-file ./install.json --hostname new-hostname

3. Using your default text editor:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --editor

  You will be able to choose from several installation examples. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the reinstallation will be run.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud baremetal reinstall ns1234.ip-11.22.33.net --editor --os debian12_64

You can visit https://eu.api.ovh.com/console/?section=%2Fdedicated%2Fserver&branch=v1#post-/dedicated/server/-serviceName-/reinstall
to see all the available parameters and real life examples.

Please note that all parameters are not compatible with all OSes.


```
ovhcloud baremetal reinstall <service_name> [flags]
```

### Options

```
      --config-drive-user-data string               Config Drive UserData
      --editor                                      Use a text editor to define parameters
      --efi-bootloader-path string                  Path of the EFI bootloader from the OS installed on the server
      --from-file string                            File containing parameters
  -h, --help                                        help for reinstall
      --hostname string                             Custom hostname
      --http-headers stringToString                 Image HTTP headers (default [])
      --image-checksum string                       Image checksum
      --image-checksum-type string                  Image checksum type
      --image-type string                           Image type (qcow, raw)
      --image-url string                            Image URL
      --init-file string                            Create a file with example parameters
      --language string                             Display language
      --os string                                   Operating system to install
      --post-installation-script string             Post-installation script
      --post-installation-script-extension string   Post-installation script extension (cmd, ps1)
      --replace                                     Replace parameters file if it already exists
      --ssh-key string                              SSH public key
      --wait                                        Wait for reinstall to be done before exiting
```

### Options inherited from parent commands

```
  -d, --debug           Activate debug mode (will log all HTTP requests details)
  -f, --format string   Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                        Examples:
                          --format 'id' (to extract a single field)
                          --format 'nested.field.subfield' (to extract a nested field)
                          --format '[id, 'name']' (to extract multiple fields as an array)
                          --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                          --format 'name+","+type' (to extract and concatenate fields in a string)
                          --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors   Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive     Interactive output
  -j, --json            Output in JSON
  -y, --yaml            Output in YAML
```

### SEE ALSO

* [ovhcloud baremetal](ovhcloud_baremetal.md)	 - Retrieve information and manage your baremetal services

