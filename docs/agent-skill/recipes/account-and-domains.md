# Recipes: account, auth & domains

## Authentication & profiles
```bash
ovhcloud login                       # interactive: region + API token
ovhcloud login --profile work        # store under a named profile
ovhcloud --profile work account get  # use a specific profile
```
Token creation page: https://api.ovh.com/createToken/ (grant `GET/POST/PUT/DELETE`
on `/*` for full access). See also [../references/commands.md](../references/commands.md).

## Account & API credentials
```bash
ovhcloud account get                          # your account information
ovhcloud account --help                       # SSH keys, API credentials, OAuth2…
ovhcloud account api oauth2 client list -o json
```

## Domains (registered names)
```bash
ovhcloud domain-name list -o json
ovhcloud domain-name get <domain>
```

## DNS zones and records
```bash
ovhcloud domain-zone list
ovhcloud domain-zone record list <zone> -o json
ovhcloud domain-zone record list <zone> --filter 'fieldType=="A"' -o json

# create / update / delete a record (confirm before delete)
ovhcloud domain-zone record create <zone> --help     # see required flags
ovhcloud domain-zone record delete <zone> <recordId>

# apply pending changes to the zone
ovhcloud domain-zone refresh <zone>
```

> As always: verify exact flags with `--help`, read results with `-o json`, and
> confirm before any `delete`/`refresh` that changes live DNS
> (see [../references/safety.md](../references/safety.md)).
