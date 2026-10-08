## ovhcloud cloud instance reinstall

Reinstall the given instance

### Synopsis

Use this command to reinstall the given instance with another image.
All data on the instance root disk will be lost.

There are two ways to choose the image:
(the following examples assume that you have already configured your default cloud project using "ovhcloud config set default_cloud_project <project_id>")

1. Using only CLI flags:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --image <image_id>

2. Using the interactive image selector:

	ovhcloud cloud instance reinstall c7e272d4-4c11-11f0-bf07-0050568ce122 --image-selector

The image must differ from the one the instance currently runs.


```
ovhcloud cloud instance reinstall <instance_id> [flags]
```

### Options

```
  -h, --help             help for reinstall
      --image string     Image to use for reinstallation
      --image-selector   Use the interactive image selector to choose the image
      --wait             Wait for reinstall to be done before exiting
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

