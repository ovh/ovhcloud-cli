# Recipes: Public Cloud (`ovhcloud cloud …`)

All Public Cloud commands need a project. Set it once:
```bash
ovhcloud cloud project list                 # find your project id
export OVH_CLOUD_PROJECT_SERVICE=<projectId> # or pass --cloud-project <id>
```

## Inspect resources
```bash
ovhcloud cloud instance list -o json
ovhcloud cloud instance list --filter 'region=="GRA11"' -o json
ovhcloud cloud instance get <instanceId>
```

## Extract an id into a variable
```bash
NET=$(ovhcloud cloud network private vrack list -o json | jq -r '.[0].id')
```

## Create and wait (async) — billed, confirm first
```bash
ovhcloud cloud network private vrack create GRA11 --name my-net --vlan-id 100 --wait
ovhcloud cloud network private vrack subnet create <networkId> \
  --name my-subnet --cidr 192.168.1.0/24 --enable-dhcp --wait
```

## Storage (S3-compatible object storage)
```bash
ovhcloud cloud storage object list -o json
ovhcloud cloud storage object create GRA --name my-bucket
```

## Delete — irreversible, confirm first
```bash
ovhcloud cloud network private vrack subnet delete <networkId> <subnetId>
ovhcloud cloud network private vrack delete <networkId>
```

> Tip: unsure which flags a `create`/`edit` takes? Run the command with
> `--help`, or `--init-file ./p.json` to generate an example body you can edit
> and pass back with `--from-file ./p.json`.
