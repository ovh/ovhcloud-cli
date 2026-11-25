## ovhcloud cloud loadbalancer pool member create

Add a member to a pool

### Synopsis

Use this command to add a member to a pool.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> --address 10.0.0.10 --protocol-port 80 --name my-member --subnet-id <subnet_id>


```
ovhcloud cloud loadbalancer pool member create <loadbalancer_id> <pool_id> [flags]
```

### Options

```
      --address string      IP address of the member
      --backup              Whether this is a backup member, only used when all the other members are down
      --editor              Use a text editor to define parameters
      --from-file string    File containing parameters
  -h, --help                help for create
      --init-file string    Create a file with example parameters
      --name string         Name of the member
      --protocol-port int   Protocol port number of the member
      --replace             Replace parameters file if it already exists
      --subnet-id string    ID of the subnet of the member
      --weight int          Weight of the member (1-256)
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

* [ovhcloud cloud loadbalancer pool member](ovhcloud_cloud_loadbalancer_pool_member.md)	 - Manage members of a loadbalancer pool

