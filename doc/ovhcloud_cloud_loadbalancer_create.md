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

	ovhcloud cloud loadbalancer create <region> --name my-lb --size SMALL --subnet-id <subnet_id>


```
ovhcloud cloud loadbalancer create <region> [flags]
```

### Options

```
      --description string   Description of the loadbalancer
      --editor               Use a text editor to define parameters
      --from-file string     File containing parameters
  -h, --help                 help for create
      --init-file string     Create a file with example parameters
      --name string          Name of the loadbalancer
      --network-id string    Network ID (optional, defaults to the network of the subnet)
      --network-ip string    IP of the loadbalancer VIP: an address inside the subnet CIDR pins the fixed address, an address outside it designates an existing floating IP to associate (requires the loadbalancer/usePublicIP IAM action)
      --replace              Replace parameters file if it already exists
      --size string          Size of the loadbalancer (SMALL, MEDIUM, LARGE, XL)
      --subnet-id string     Subnet ID
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string         Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud cloud loadbalancer](ovhcloud_cloud_loadbalancer.md)	 - Manage loadbalancers in the given cloud project

