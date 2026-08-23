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
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud telephony](ovhcloud_telephony.md)	 - Retrieve information and manage your Telephony services

