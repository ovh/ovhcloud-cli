## ovhcloud email-mxplan edit

Edit the given Email MXPlan

```
ovhcloud email-mxplan edit <service_name> [flags]
```

### Options

```
      --complexity-enabled               Enable policy for strong and secure passwords
      --display-name string              Service displayName
      --editor                           Use a text editor to define parameters
  -h, --help                             help for edit
      --lockout-duration int             Number of minutes account will remain locked if it occurs
      --lockout-observation-window int   Number of minutes that must elapse after a failed logon to reset lockout trigger
      --lockout-threshold int            Number of attempts before account to be locked
      --max-password-age int             Maximum number of days that account's password is valid before expiration
      --max-receive-size int             Maximum message size that you can receive in MB
      --max-send-size int                Maximum message size that you can send in MB
      --min-password-age int             Minimum number of days before able to change account's password
      --min-password-length int          Minimum number of characters password must contain
      --spam-check-dkim                  Check DKIM of message
      --spam-check-spf                   Check SPF of message
      --spam-delete-spam                 If message is a spam delete it
      --spam-delete-virus                If message is a virus delete it
      --spam-put-in-junk                 If message is a spam or virus put in junk. Overridden by deleteSpam or deleteVirus
      --spam-tag-spam                    If message is a spam change its subject
      --spam-tag-virus                   If message is a virus change its subject
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

* [ovhcloud email-mxplan](ovhcloud_email-mxplan.md)	 - Retrieve information and manage your Email MXPlan services

