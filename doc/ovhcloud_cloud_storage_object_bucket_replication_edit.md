## ovhcloud cloud storage object bucket replication edit

Replace the replication rules of the given bucket

### Synopsis

Use this command to replace all the replication rules of the given bucket.
There are two ways to define the new rules:

1. Using a configuration file:

  First you can generate an example of rules file using the following command:

	ovhcloud cloud storage object bucket replication edit --init-file ./rules.json

  After editing the file, run:

	ovhcloud cloud storage object bucket replication edit GRA_my-bucket --from-file ./rules.json

  The rules of the file replace all the current rules of the bucket.

2. Using your default text editor, starting from the current rules:

	ovhcloud cloud storage object bucket replication edit GRA_my-bucket --editor


```
ovhcloud cloud storage object bucket replication edit <bucket_id> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from-file string   File containing parameters
  -h, --help               help for edit
      --init-file string   Create a file with example parameters
      --replace            Replace parameters file if it already exists
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

* [ovhcloud cloud storage object bucket replication](ovhcloud_cloud_storage_object_bucket_replication.md)	 - Manage the replication rules of S3™* compatible buckets (* S3 is a trademark filed by Amazon Technologies,Inc. OVHcloud's service is not sponsored by, endorsed by, or otherwise affiliated with Amazon Technologies,Inc.)

