## ovhcloud webhosting email-option

Manage email options

### Options

```
  -h, --help   help for email-option
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
                         
                         When extracting a single scalar field, the value is printed without surrounding
                         quotes (useful for scripting); objects and arrays are still rendered as JSON.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Retrieve information and manage your WebHosting services
* [ovhcloud webhosting email-option get](ovhcloud_webhosting_email-option_get.md)	 - Get an email option
* [ovhcloud webhosting email-option list](ovhcloud_webhosting_email-option_list.md)	 - List email options
* [ovhcloud webhosting email-option service-info](ovhcloud_webhosting_email-option_service-info.md)	 - Get email option service info
* [ovhcloud webhosting email-option terminate](ovhcloud_webhosting_email-option_terminate.md)	 - Terminate an email option

