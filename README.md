# pulumi-ohdear

A Pulumi provider for [Oh Dear](https://ohdear.app), built with
[pulumi-go-provider](https://github.com/pulumi/pulumi-go-provider) (infer SDK).

## Layout

| Path | Purpose |
| --- | --- |
| `provider/provider.go` | Provider registration (`infer.NewProviderBuilder`) |
| `provider/config.go` | Provider config — `apiToken` / `OHDEAR_API_TOKEN` |
| `provider/site.go` | `ohdear:index:Site` resource (stub CRUD) |
| `provider/cmd/pulumi-resource-ohdear/main.go` | Plugin entrypoint |

## Develop

```sh
make build      # -> bin/pulumi-resource-ohdear
make install    # copy binary onto $GOPATH/bin
make schema     # -> schema.json
make sdk        # -> sdk/nodejs
```

`make install` then `pulumi package add ./bin/pulumi-resource-ohdear` in a stack to try it.

## Status

Boilerplate only. `Site` echoes its inputs and mints a placeholder ID; no Oh Dear
API calls yet. Next: add an API client and real Create/Read/Update/Delete/Diff.
