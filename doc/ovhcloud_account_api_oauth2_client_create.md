## ovhcloud account api oauth2 client create

Create a new OAuth2 client

```
ovhcloud account api oauth2 client create [flags]
```

### Options

```
      --callback-urls stringArray   Callback URLs for the OAuth2 client
      --description string          Description of the OAuth2 client
      --editor                      Use a text editor to define parameters
      --flow string                 OAuth2 flow type (default: AUTHORIZATION_CODE) (default "AUTHORIZATION_CODE")
      --from-file string            File containing parameters
  -h, --help                        help for create
      --init-file string            Create a file with example parameters
      --name string                 Name of the OAuth2 client
      --replace                     Replace parameters file if it already exists
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud account api oauth2 client](ovhcloud_account_api_oauth2_client.md)	 - Manage your OAuth2 clients

