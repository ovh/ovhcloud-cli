# Product: Network & vRack

- `ovhcloud cloud network …` — Public Cloud networking (needs a project:
  `--cloud-project <id>` / `OVH_CLOUD_PROJECT_SERVICE`). Sub-resources:
  `private` (vRack private networks + `subnet`), `public`, `gateway`.
- `ovhcloud vrack …` — account-level vRack services (`list`/`get`/`edit`).

```bash
# Public Cloud private networks (API v2)
ovhcloud cloud network private vrack list -o json
ovhcloud cloud network private vrack create <region> --name my-net --wait
ovhcloud cloud network private vrack subnet create <networkId> --region <r> \
  --name my-subnet --cidr 192.168.1.0/24 --enable-dhcp --wait
ovhcloud cloud network gateway list -o json

# vRack (account level)
ovhcloud vrack list -o json
ovhcloud vrack get <serviceName>
```

> Gateways and some network resources are **billed**; `delete` is destructive.
> Verify flags with `--help` (see [../references/safety.md](../references/safety.md)).
