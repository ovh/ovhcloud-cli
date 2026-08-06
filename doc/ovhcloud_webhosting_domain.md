## ovhcloud webhosting domain

Manage attached domains

### Options

```
  -h, --help   help for domain
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
  -y, --yes              Skip confirmation prompts (assume yes); required to delete non-interactively
```

### SEE ALSO

* [ovhcloud webhosting](ovhcloud_webhosting.md)	 - Manage Web Hosting (databases, cron, SSL, env vars, logs)
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

