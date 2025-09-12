## ovhcloud cloud kube kubeconfig reset

Reset the kubeconfig for the given Kubernetes cluster. Certificates will be regenerated and nodes will be reinstalled

```
ovhcloud cloud kube kubeconfig reset <cluster_id> [flags]
```

### Options

```
  -h, --help   help for reset
```

### Options inherited from parent commands

```
      --cloud-project string   Cloud project ID
  -d, --debug                  Activate debug mode (will log all HTTP requests details)
  -f, --format string          Output value according to given format (expression using gval format)
  -e, --ignore-errors          Ignore errors in API calls when it is not fatal to the execution
  -i, --interactive            Interactive output
  -j, --json                   Output in JSON
  -y, --yaml                   Output in YAML
```

### SEE ALSO

* [ovhcloud cloud kube kubeconfig](ovhcloud_cloud_kube_kubeconfig.md)	 - Manage the kubeconfig for the given Kubernetes cluster

