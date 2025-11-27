## ovhcloud webhosting email-option

Manage email options

```
  ovhcloud webhosting email-option [command]
```

### Options

```
  -h, --help   help for email-option
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

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting email-option get](ovhcloud_webhosting_email-option_get.md)	 - Get an email option
* [ovhcloud webhosting email-option list](ovhcloud_webhosting_email-option_list.md)	 - List email options
* [ovhcloud webhosting email-option service-info](ovhcloud_webhosting_email-option_service-info.md)	 - Get email option service info
* [ovhcloud webhosting email-option terminate](ovhcloud_webhosting_email-option_terminate.md)	 - Terminate an email option
