## ovhcloud cloud key-manager secret create

Create a new Key Manager secret

```
ovhcloud cloud key-manager secret create [flags]
```

### Options

```
      --algorithm string              Algorithm associated with the secret (AES, DH, DSA, EC, RSA)
      --availability-zone string      Availability zone within the region
      --bit-length int                Bit length of the secret (128, 256, 512, 1024, 2048, 4096)
      --editor                        Use a text editor to define parameters
      --expiration string             Expiration date and time of the secret (RFC3339)
      --from-file string              File containing parameters
  -h, --help                          help for create
      --init-file string              Create a file with example parameters
      --metadata stringToString       Metadata key-value pairs for the secret (default [])
      --mode string                   Mode of the algorithm (CBC, CTR)
      --name string                   Human-readable name of the secret
      --payload string                Secret payload data (base64-encoded, write-only). Requires --payload-content-type
      --payload-content-type string   Content type of the payload (APPLICATION_OCTET_STREAM, APPLICATION_PKCS8, APPLICATION_PKIX_CERT, TEXT_PLAIN)
      --region string                 Region code where the secret is located
      --replace                       Replace parameters file if it already exists
      --secret-type string            Type of the secret (CERTIFICATE, OPAQUE, PASSPHRASE, PRIVATE, PUBLIC, SYMMETRIC)
      --wait                          Wait for the secret to be ready before exiting
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -o, --output string          Output format: json, yaml, interactive, or a custom format expression (using https://github.com/PaesslerAG/gval syntax)
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
      --profile string         Use a specific profile from the configuration file
      --raw                    Output the extracted value without JSON quoting (use with -o '<field>'), useful for scripting
                               Example:
                                 --output 'id' --raw   (prints the id without surrounding quotes)
```

### SEE ALSO

* [ovhcloud cloud key-manager secret](ovhcloud_cloud_key-manager_secret.md)	 - Manage Key Manager secrets

