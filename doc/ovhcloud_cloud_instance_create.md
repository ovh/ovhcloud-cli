## ovhcloud cloud instance create

Create a new instance

### Synopsis

Use this command to create an instance in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud instance create GRA9 --name MyNewInstance --boot-from.image <image_id> --flavor <flavor_id>

  To boot from an existing bootable block storage volume instead of an image:

	ovhcloud cloud instance create GRA9 --name MyNewInstance --boot-from.volume <volume_id> --flavor <flavor_id>

2. Using a configuration file:

  First you can generate an example of installation file using the following command:

	ovhcloud cloud instance create BHS5 --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud instance create BHS5 --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud instance create BHS5

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud instance create GRA11 --from-file ./params.json --name NameOverriden

  Network and volume flags (--network.*, --boot-from.volume) add entries to the networks and volumes of the file
  instead of replacing them.

  It is also possible to use the interactive image and flavor selector to define the image and flavor parameters, like the following:

	ovhcloud cloud instance create BHS5 --init-file ./params.json --image-selector --flavor-selector

3. Using your default text editor:

	ovhcloud cloud instance create GRA11 --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud instance create RBX8 --editor --flavor <flavor_id>

  You can also use the interactive image and flavor selector to define the image and flavor parameters, like the following:

	ovhcloud cloud instance create RBX8 --editor --image-selector --flavor-selector


```
ovhcloud cloud instance create <region (e.g. GRA9, BHS5, SBG3)> [flags]
```

### Options

```
      --availability-zone string           Availability zone
      --boot-from.image string             Image ID to boot from (you can use 'ovhcloud cloud instance image list' to get the image ID or 'ovhcloud cloud instance snapshot ls' to get the snapshots)
      --boot-from.volume string            ID of an existing bootable block storage volume to boot from, instead of an image (the volume must be available and in the same region; it is kept when the instance is deleted)
      --editor                             Use a text editor to define parameters
      --flavor string                      Flavor ID (you can use 'ovhcloud cloud instance flavor list' to get the flavor ID)
      --flavor-selector                    Use the interactive flavor selector
      --from-file string                   File containing parameters
      --group string                       Group ID
  -h, --help                               help for create
      --image-selector                     Use the interactive image selector
      --init-file string                   Create a file with example parameters
      --name string                        Instance name
      --network.private.id string          ID of the existing private network
      --network.private.ip string          Instance IP in the private network, or the address of an existing floating IP to associate
      --network.private.subnet-id string   Existing subnet ID
      --network.public                     Attach the new instance to the public network
      --replace                            Replace parameters file if it already exists
      --ssh-key.name string                Existing SSH key name (you can use 'ovhcloud cloud ssh-key create' to create one)
      --user-data string                   Configuration information or scripts to use upon launch
      --wait                               Wait for instance creation to be done before exiting
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

* [ovhcloud cloud instance](ovhcloud_cloud_instance.md)	 - Manage instances in the given cloud project

