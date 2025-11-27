## ovhcloud webhosting email

Manage automated emails

```
  ovhcloud webhosting email [command]
```

### Options

```
  -h, --help   help for email
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
* [ovhcloud webhosting email bounces](ovhcloud_webhosting_email_bounces.md)	 - List recent email bounces
* [ovhcloud webhosting email info](ovhcloud_webhosting_email_info.md)	 - Get email sending settings
* [ovhcloud webhosting email request-action](ovhcloud_webhosting_email_request-action.md)	 - Request an email action
* [ovhcloud webhosting email update](ovhcloud_webhosting_email_update.md)	 - Update email sending settings
* [ovhcloud webhosting email volumes](ovhcloud_webhosting_email_volumes.md)	 - List email sending volumes
