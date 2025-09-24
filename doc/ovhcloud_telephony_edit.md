## ovhcloud telephony edit

Edit the given Telephony service

```
ovhcloud telephony edit <service_name> [flags]
```

### Options

```
      --credit-threshold-currency string   Currency code (AUD, CAD, CZK, EUR, GBP, INR, LTL, MAD, N/A, PLN, SGD, TND, USD, XOF, points)
      --credit-threshold-text string       Text for credit threshold
      --credit-threshold-value int         Value for credit threshold
      --description string                 Description of service
      --editor                             Use a text editor to define parameters
  -h, --help                               help for edit
      --hidden-external-number             Hide called numbers in end-of-month call details CSV
      --override-displayed-number          Override number displayed for calls between services of your billing account
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

* [ovhcloud telephony](ovhcloud_telephony.md)	 - Retrieve information and manage your Telephony services

