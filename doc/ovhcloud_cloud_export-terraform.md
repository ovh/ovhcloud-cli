## ovhcloud cloud export-terraform

Export the project's resources as Terraform import blocks

### Synopsis

Generate Terraform "import" blocks for the resources of a Public Cloud project,
to adopt Infrastructure-as-Code on an existing project without retyping everything.

The command writes an "imports.tf" file; Terraform then generates the matching
configuration for you:

  ovhcloud cloud export-terraform --cloud-project <id>
  terraform plan -generate-config-out=generated.tf

Only resources supported by the OVHcloud Terraform provider are exported. This
command reads your infrastructure but never modifies it.

```
ovhcloud cloud export-terraform [flags]
```

### Examples

```
  # Export all supported resources of a project
  ovhcloud cloud export-terraform --cloud-project <project_id> --output-dir ./tf

  # Restrict to some resource types
  ovhcloud cloud export-terraform --cloud-project <project_id> --resources network,user
```

### Options

```
      --cloud-project string   Cloud project ID
  -h, --help                   help for export-terraform
      --output-dir string      Directory where imports.tf is written (default ".")
      --resources strings      Restrict export to these resource types (e.g. network,user)
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

* [ovhcloud cloud](ovhcloud_cloud.md)	 - Manage your projects and services in the Public Cloud universe (MKS, MPR, MRS, Object Storage...)

