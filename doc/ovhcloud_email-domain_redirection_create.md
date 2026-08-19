## ovhcloud email-domain redirection create

Create a new email redirection

```
ovhcloud email-domain redirection create <service_name> [flags]
```

### Options

```
      --editor             Use a text editor to define parameters
      --from string        Source email address (e.g., alias@domain.com)
      --from-file string   File containing parameters
  -h, --help               help for create
      --init-file string   Create a file with example parameters
      --local-copy         Keep a local copy of the email
      --replace            Replace parameters file if it already exists
      --to string          Destination email address
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud email-domain redirection](ovhcloud_email-domain_redirection.md)	 - Manage email redirections for your domain

