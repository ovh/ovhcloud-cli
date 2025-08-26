## ovhcloud cloud instance create

Create a new instance

### Synopsis

Use this command to create an instance in the given public cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

  ovhcloud cloud instance create GRA9 --name MyNewInstance --boot-from.image <image_id> --flavor <flavor_id> ...

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
ovhcloud cloud instance create <region (e.g. GRA, BHS, SBG)> [flags]
```

### Options

```
      --availability-zone string                                Availability zone
      --backup-cron string                                      Autobackup Unix Cron pattern (eg: '0 0 * * *')
      --backup-rotation int                                     Number of backups to keep
      --billing-period string                                   Billing period (hourly, monthly), default is hourly (default "hourly")
      --boot-from.image string                                  Image ID to boot from
      --boot-from.volume string                                 Volume ID to boot from
      --bulk int                                                Number of instances to create
      --editor                                                  Use a text editor to define parameters
      --flavor string                                           Flavor ID
      --flavor-selector                                         Use the interactive flavor selector
      --from-file string                                        File containing parameters
      --group string                                            Group ID
  -h, --help                                                    help for create
      --image-selector                                          Use the interactive image selector
      --init-file string                                        Create a file with example parameters
      --name string                                             Instance name
      --network.private.create.name string                      Name for the private network to create
      --network.private.create.subnet-cidr string               CIDR for the subnet to create
      --network.private.create.subnet-enable-dhcp               Enable DHCP for the subnet to create
      --network.private.create.subnet-ip-version int            IP version for the subnet to create
      --network.private.create.vlan-id int                      VLAN ID for the private network to create
      --network.private.floating-ip.create.description string   Description for the floating IP to create
      --network.private.floating-ip.id string                   ID of an existing floating IP
      --network.private.gateway.create.model string             Model for the gateway to create (s, m, l)
      --network.private.gateway.create.name string              Name for the gateway to create
      --network.private.gateway.id string                       ID of the existing gateway to attach to the private network
      --network.private.id string                               ID of the existing private network
      --network.private.ip string                               Instance IP in the private network
      --network.private.subnet-id string                        Existing subnet ID
      --network.public                                          Set the new instance as public
      --replace                                                 Replace parameters file if it already exists
      --ssh-key.create.name string                              Name for the SSH key to create
      --ssh-key.create.public-key string                        Public key for the SSH key to create
      --ssh-key.name string                                     Existing SSH key name
      --user-data string                                        Configuration information or scripts to use upon launch
      --wait                                                    Wait for instance creation to be done before exiting
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

###### Auto generated by spf13/cobra on 26-Aug-2025
