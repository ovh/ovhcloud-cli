## ovhcloud vps reinstall

Reinstall the given VPS

```
ovhcloud vps reinstall <service_name> [flags]
```

### Options

```
      --do-not-send-password                   Do not send the new password after reinstallation (only if sshKey defined)
      --editor                                 Use a text editor to define parameters
      --from-file string                       File containing parameters
  -h, --help                                   help for reinstall
      --image-id string                        ID of the image to use for reinstallation
      --image-selector                         Use the interactive image selector
      --init-file string                       Create a file with example parameters
      --install-rtm                            Install RTM during reinstallation
      --public-ssh-key string                  Public SSH key to pre-install on your VPS
      --replace                                Replace parameters file if it already exists
      --ssh-key ovh-cli account ssh-key list   SSH key name to pre-install on your VPS (name can be found running ovh-cli account ssh-key list)
      --ssh-key-selector                       Use the interactive SSH key selector
      --wait                                   Wait for reinstall to be done before exiting
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

* [ovhcloud vps](ovhcloud_vps.md)	 - Retrieve information and manage your VPS services

