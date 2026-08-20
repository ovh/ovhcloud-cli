## ovhcloud account bill

Read your invoices

### Options

```
  -h, --help   help for bill
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

* [ovhcloud account](ovhcloud_account.md)	 - Manage your account
* [ovhcloud account bill details](ovhcloud_account_bill_details.md)	 - List what one invoice charges for
* [ovhcloud account bill get](ovhcloud_account_bill_get.md)	 - Get one invoice
* [ovhcloud account bill list](ovhcloud_account_bill_list.md)	 - List your invoices

