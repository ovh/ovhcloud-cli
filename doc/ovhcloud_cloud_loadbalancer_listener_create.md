## ovhcloud cloud loadbalancer listener create

Create a listener on a loadbalancer

### Synopsis

Use this command to create a listener, optionally with its default pool.
There are three ways to define the parameters:

1. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --init-file ./params.json

  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --from-file ./params.json

  Note that you can also pipe the content of the parameters file.

2. Using your default text editor:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --editor

3. Using only CLI flags:

	ovhcloud cloud loadbalancer listener create <loadbalancer_id> --name my-listener --protocol HTTP --port 80 --default-pool-id <pool_id>


```
ovhcloud cloud loadbalancer listener create <loadbalancer_id> [flags]
```

### Options

```
      --default-pool-id string             Default pool ID
      --default-tls-container-ref string   Reference of the default TLS certificate container
      --description string                 Description of the listener
      --editor                             Use a text editor to define parameters
      --from-file string                   File containing parameters
  -h, --help                               help for create
      --init-file string                   Create a file with example parameters
      --name string                        Name of the listener
      --port int                           Port to listen on
      --protocol string                    Protocol (HTTP, HTTPS, TERMINATED_HTTPS, TCP, UDP, SCTP)
      --replace                            Replace parameters file if it already exists
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

* [ovhcloud cloud loadbalancer listener](ovhcloud_cloud_loadbalancer_listener.md)	 - Manage listeners of loadbalancers

