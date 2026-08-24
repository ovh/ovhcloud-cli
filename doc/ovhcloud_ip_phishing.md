## ovhcloud ip phishing

Read the phishing URLs reported on the given IP block

### Options

```
  -h, --help   help for phishing
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
* [ovhcloud ip phishing get](ovhcloud_ip_phishing_get.md)	 - Show one phishing entry
* [ovhcloud ip phishing list](ovhcloud_ip_phishing_list.md)	 - List the phishing entries reported on this block

