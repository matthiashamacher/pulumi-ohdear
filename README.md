# pulumi-ohdear

A Pulumi provider for [Oh Dear](https://ohdear.app), built with
[pulumi-go-provider](https://github.com/pulumi/pulumi-go-provider) (infer SDK).

## Layout

| Path | Purpose |
| --- | --- |
| `provider/provider.go` | Provider registration (`infer.NewProviderBuilder`) |
| `provider/config.go` | Provider config — `apiToken` / `OHDEAR_API_TOKEN` |
| `provider/site.go` | `ohdear:index:Site` resource (stub CRUD) |
| `provider/tag.go` | `ohdear:index:Tag` — create + read (no update/delete in the API) |
| `provider/taggroup.go` | `ohdear:index:TagGroup` — full CRUD |
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

`Tag` and `TagGroup` call the real API. `Tag` has no update or delete endpoint
upstream, so input changes replace it and destroy only drops it from state.
`Site` is still a stub. Resource tokens are mapped to the `index` module.
