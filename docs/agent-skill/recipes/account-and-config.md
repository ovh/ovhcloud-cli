# Recipes: account, auth & configuration

## Authentication & profiles
```bash
ovhcloud login                       # interactive: region + API token
ovhcloud login --profile work        # store under a named profile
ovhcloud --profile work account get  # use a specific profile
ovhcloud logout                      # revoke credentials and remove them from the config
```
Token creation page: https://api.ovh.com/createToken/ (grant `GET/POST/PUT/DELETE`
on `/*` for full access). See also [../references/commands.md](../references/commands.md).

## Account & API credentials
```bash
ovhcloud account get                          # your account information
ovhcloud account --help                       # SSH keys, API credentials, OAuth2…
ovhcloud account api oauth2 client list -o json
```

## Configuration
```bash
ovhcloud config show                          # display the current CLI configuration
ovhcloud config set <key> <value>             # set a configuration value
ovhcloud config set-endpoint EU               # target an API endpoint (EU, CA, US) or a full URL
ovhcloud config profile --help                # manage profiles for multiple accounts
```

> As always: verify exact flags with `--help` and read results with `-o json`
> (see [../references/safety.md](../references/safety.md)).
