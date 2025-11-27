## ovhcloud webhosting domain

Manage attached domains

```
  ovhcloud webhosting domain [command]
```

### Options

```
  -h, --help   help for domain
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
* [ovhcloud webhosting domain add](ovhcloud_webhosting_domain_add.md)	 - Attach a domain
* [ovhcloud webhosting domain available-offer](ovhcloud_webhosting_domain_available-offer.md)	 - List hosting offers available for a domain
* [ovhcloud webhosting domain delete](ovhcloud_webhosting_domain_delete.md)	 - Delete an attached domain
* [ovhcloud webhosting domain dig-status](ovhcloud_webhosting_domain_dig-status.md)	 - Check DNS status for an attached domain
* [ovhcloud webhosting domain find](ovhcloud_webhosting_domain_find.md)	 - Find hosting service linked to a domain
* [ovhcloud webhosting domain get](ovhcloud_webhosting_domain_get.md)	 - Get an attached domain
* [ovhcloud webhosting domain list](ovhcloud_webhosting_domain_list.md)	 - List attached domains
* [ovhcloud webhosting domain purge-cache](ovhcloud_webhosting_domain_purge-cache.md)	 - Purge CDN cache for an attached domain
* [ovhcloud webhosting domain restart](ovhcloud_webhosting_domain_restart.md)	 - Restart an attached domain
* [ovhcloud webhosting domain update](ovhcloud_webhosting_domain_update.md)	 - Update an attached domain
