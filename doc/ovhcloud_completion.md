## ovhcloud completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for ovhcloud CLI.

To load completions in your current shell session:

  bash:
    source <(ovhcloud completion bash)

  zsh:
    source <(ovhcloud completion zsh)

  fish:
    ovhcloud completion fish | source

To make completions permanent, run:

  ovhcloud completion install


```
ovhcloud completion [bash|zsh|fish|powershell] [flags]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
  -d, --debug            Activate debug mode (will log all HTTP requests details)
  -e, --ignore-errors    Ignore errors in API calls when it is not fatal to the execution
  -o, --output string    Output format: json, yaml, interactive, or a custom format expression. Run 'ovhcloud --help' for the full list with examples.
      --profile string   Use a specific profile from the configuration file
```

### SEE ALSO

* [ovhcloud](ovhcloud.md)	 - CLI to manage your OVHcloud services
* [ovhcloud completion install](ovhcloud_completion_install.md)	 - Install shell completion permanently in your shell rc file

