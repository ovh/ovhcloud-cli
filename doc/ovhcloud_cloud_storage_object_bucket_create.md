## ovhcloud cloud storage object bucket create

Create a new S3™* compatible bucket (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

### Synopsis

Use this command to create a S3™* compatible bucket in the given cloud project.
The region is always taken from the <region> argument.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud cloud storage object bucket create BHS --name my-new-bucket

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud cloud storage object bucket create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud cloud storage object bucket create GRA --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud cloud storage object bucket create GRA

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud cloud storage object bucket create GRA --from-file ./params.json --name name-overridden

3. Using your default text editor:

	ovhcloud cloud storage object bucket create GRA --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud cloud storage object bucket create GRA --editor --name name-overridden

*S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.


```
ovhcloud cloud storage object bucket create <region> [flags]
```

### Options

```
      --editor                           Use a text editor to define parameters
      --encryption-algorithm string      Server-side encryption algorithm (AES256, PLAINTEXT)
      --from-file string                 File containing parameters
  -h, --help                             help for create
      --init-file string                 Create a file with example parameters
      --name string                      Name of the bucket (must be globally unique and DNS-compatible)
      --object-lock-mode string          Object lock retention mode (COMPLIANCE, GOVERNANCE), requires versioning to be enabled
      --object-lock-retention-days int   Number of days to retain objects
      --owner-user-id string             Owner user ID of the bucket
      --replace                          Replace parameters file if it already exists
      --tag stringToString               Bucket tags as key=value pairs (default [])
      --versioning-status string         Versioning status (DISABLED, ENABLED, SUSPENDED)
      --wait                             Wait for the bucket to be ready before exiting
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

* [ovhcloud cloud storage object bucket](ovhcloud_cloud_storage_object_bucket.md)	 - Manage object storage buckets in the given cloud project

