## ovhcloud webhosting cdn domain

Manage CDN domains

```
  ovhcloud webhosting cdn domain [command]
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

* [ovhcloud webhosting cdn](ovhcloud_webhosting_cdn.md)	 - Manage CDN
* [ovhcloud webhosting cdn domain get](ovhcloud_webhosting_cdn_domain_get.md)	 - Get a CDN domain
* [ovhcloud webhosting cdn domain list](ovhcloud_webhosting_cdn_domain_list.md)	 - List CDN domains
* [ovhcloud webhosting cdn domain option](ovhcloud_webhosting_cdn_domain_option.md)	 - Manage CDN domain options
* [ovhcloud webhosting cdn domain purge](ovhcloud_webhosting_cdn_domain_purge.md)	 - Purge CDN domain cache
* [ovhcloud webhosting cdn domain refresh](ovhcloud_webhosting_cdn_domain_refresh.md)	 - Refresh CDN domain
* [ovhcloud webhosting cdn domain statistics](ovhcloud_webhosting_cdn_domain_statistics.md)	 - Get CDN domain statistics
