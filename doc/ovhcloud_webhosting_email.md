## ovhcloud webhosting email

Manage automated emails

### Options

```
  -h, --help   help for email
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
* [ovhcloud webhosting email bounces](ovhcloud_webhosting_email_bounces.md)	 - List recent email bounces
* [ovhcloud webhosting email info](ovhcloud_webhosting_email_info.md)	 - Get email sending settings
* [ovhcloud webhosting email request-action](ovhcloud_webhosting_email_request-action.md)	 - Request an email action
* [ovhcloud webhosting email update](ovhcloud_webhosting_email_update.md)	 - Update email sending settings
* [ovhcloud webhosting email volumes](ovhcloud_webhosting_email_volumes.md)	 - List email sending volumes

