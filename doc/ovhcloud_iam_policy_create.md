## ovhcloud iam policy create

Create a new policy

### Synopsis

Use this command to create a new policy.
There are three ways to define the creation parameters:

1. Using only CLI flags:

	ovhcloud iam policy create --name MyPolicy --allow 'domain:apiovh:get' --identity 'urn:v1:eu:identity:account:aa1-ovh' --resource 'urn:v1:eu:resource:domain:*'

2. Using a configuration file:

  First you can generate an example of parameters file using the following command:

	ovhcloud iam policy create --init-file ./params.json

  You will be able to choose from several examples of parameters. Once an example has been selected, the content is written in the given file.
  After editing the file to set the correct creation parameters, run:

	ovhcloud iam policy create --from-file ./params.json

  Note that you can also pipe the content of the parameters file, like the following:

	cat ./params.json | ovhcloud iam policy create

  In both cases, you can override the parameters in the given file using command line flags, for example:

	ovhcloud iam policy create --from-file ./params.json --name MyPolicy --allow 'domain:apiovh:get' --identity 'urn:v1:eu:identity:account:aa1-ovh' --resource 'urn:v1:eu:resource:domain:*'

3. Using your default text editor:

	ovhcloud iam policy create --editor

  You will be able to choose from several examples of parameters. Once an example has been selected, the CLI will open your
  default text editor to update the parameters. When saving the file, the creation will start.

  Note that it is also possible to override values in the presented examples using command line flags like the following:

	ovhcloud iam policy create --editor --name MyPolicy --allow 'domain:apiovh:get' --identity 'urn:v1:eu:identity:account:aa1-ovh' --resource 'urn:v1:eu:resource:domain:*'


```
ovhcloud iam policy create [flags]
```

### Options

```
      --allow strings               List of allowed actions
      --deny strings                List of denied actions
      --description string          Description of the policy
      --editor                      Use a text editor to define parameters
      --except strings              List of actions to filter from the allowed list
      --expiredAt string            Expiration date of the policy (RFC3339 format), after this date it will no longer be applied
      --from-file string            File containing parameters
  -h, --help                        help for create
      --identity strings            Identities to which the policy applies
      --init-file string            Create a file with example parameters
      --name string                 Name of the policy
      --permissions-group strings   Permissions group URNs
      --replace                     Replace parameters file if it already exists
      --resource strings            Resource URNs
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

* [ovhcloud iam policy](ovhcloud_iam_policy.md)	 - Manage IAM policies

