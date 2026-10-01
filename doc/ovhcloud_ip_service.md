## ovhcloud ip service

Manage your IP services: contacts, renewal and termination

### Options

```
  -h, --help   help for service
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
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud ip](ovhcloud_ip.md)	 - Retrieve information and manage your IP services
* [ovhcloud ip service change-contact](ovhcloud_ip_service_change-contact.md)	 - Start a contact change procedure on an IP service
* [ovhcloud ip service confirm-termination](ovhcloud_ip_service_confirm-termination.md)	 - Confirm the termination of an IP service with the emailed token
* [ovhcloud ip service edit](ovhcloud_ip_service_edit.md)	 - Edit an IP service
* [ovhcloud ip service get](ovhcloud_ip_service_get.md)	 - Show one IP service
* [ovhcloud ip service list](ovhcloud_ip_service_list.md)	 - List your IP services
* [ovhcloud ip service service-info](ovhcloud_ip_service_service-info.md)	 - Manage the billing information of an IP service
* [ovhcloud ip service terminate](ovhcloud_ip_service_terminate.md)	 - Ask for the termination of an IP service

