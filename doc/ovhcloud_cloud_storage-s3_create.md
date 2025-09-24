## ovhcloud cloud storage-s3 create

Create a new S3™* compatible storage container (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

### Synopsis

Use this command to create a S3™* compatible storage container in the given cloud project.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud storage-s3 create BHS --name mynewContainer

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud storage-s3 create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud storage-s3 create GRA --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud storage-s3 create GRA

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud storage-s3 create GRA --from-file ./params.json --name nameoverriden

3. Using your default text editor:

	ovhcloud cloud storage-s3 create GRA --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud storage-s3 create GRA --editor --name nameoverriden

*S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.


```
ovhcloud cloud storage-s3 create <region> [flags]
```

### Options

```
      --editor                            Use a text editor to define parameters
      --encryption-sse-algorithm string   Encryption SSE Algorithm (AES256, plaintext)
      --from-file string                  File containing parameters
  -h, --help                              help for create
      --init-file string                  Create a file with example parameters
      --name string                       Name of the storage container
      --object-lock-rule-mode string      Object lock mode (compliance, governance)
      --object-lock-rule-period string    Object lock period (e.g., P3Y6M4DT12H30M5S)
      --object-lock-status string         Object lock status (disabled, enabled)
      --owner-id int                      Owner ID of the storage container
      --replace                           Replace parameters file if it already exists
      --tag stringToString                Container tags as key=value pairs (default [])
      --versioning-status string          Versioning status (disabled, enabled, suspended)
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using https://github.com/PaesslerAG/gval syntax)
                               Examples:
                                 --format 'id' (to extract a single field)
                                 --format 'nested.field.subfield' (to extract a nested field)
                                 --format '[id, 'name']' (to extract multiple fields as an array)
                                 --format '{"newKey": oldKey, "otherKey": nested.field}' (to extract and rename fields in an object)
                                 --format 'name+","+type' (to extract and concatenate fields in a string)
                                 --format '(nbFieldA + nbFieldB) * 10' (to compute values from numeric fields)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud storage-s3](ovhcloud_cloud_storage-s3.md)	 - Manage S3™* compatible storage containers in the given cloud project (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

