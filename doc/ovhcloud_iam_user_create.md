## ovhcloud iam user create

Create a new user

### Synopsis

Use this command to create a new IAM user.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud iam user create --login my_user --password 'MyStrongPassword123!' --email fake.email@ovhcloud.com

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud iam user create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud iam user create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud iam user create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud iam user create --from-file ./params.json --login nameoverriden

3. Using your default text editor:

	ovhcloud iam user create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud iam user create --editor --login nameoverriden


```
ovhcloud iam user create [flags]
```

### Options

```
      --description string   Description of the user
      --editor               Use a text editor to define parameters
      --email string         Email of the user
      --from-file string     File containing parameters
      --group string         Group of the user
  -h, --help                 help for create
      --init-file string     Create a file with example parameters
      --login string         Login of the user
      --password string      Password of the user
      --replace              Replace parameters file if it already exists
      --type string          Type of the user (ROOT, SERVICE, USER)
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

* [ovhcloud iam user](ovhcloud_iam_user.md)	 - Manage IAM users

