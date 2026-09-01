# pulumi-ohdear

A Pulumi provider for [Oh Dear](https://ohdear.app), built with
[pulumi-go-provider](https://github.com/pulumi/pulumi-go-provider) (infer SDK).

This is an independent, community-maintained project, not an official Oh Dear
package. The Oh Dear team was kind enough to let it use their logo — with
thanks, and the same disclaimer: this isn't built or supported by them.

## Layout

| Path | Purpose |
| --- | --- |
| `provider/provider.go` | Provider registration (`infer.NewProviderBuilder`) |
| `provider/config.go` | Provider config — `apiToken` / `OHDEAR_API_TOKEN` |
| `provider/site.go` | `ohdear:index:Site` resource (stub CRUD) |
| `provider/monitor.go` | `ohdear:index:Monitor` — full CRUD (top-level fields only) |
| `provider/croncheck.go` | `ohdear:index:CronCheck` — full CRUD; cron job monitoring on a monitor |
| `provider/recurringmaintenanceperiod.go` | `ohdear:index:RecurringMaintenancePeriod` — full CRUD; recurring maintenance schedule |
| `provider/notificationdestination.go` | `ohdear:index:NotificationDestination` — full CRUD for team/monitor/tag/tagGroup levels |
| `provider/statuspage.go` | `ohdear:index:StatusPage` — CRUD; title/team replace, monitors sync in place |
| `provider/tag.go` | `ohdear:index:Tag` — create + read (no update/delete in the API) |
| `provider/taggroup.go` | `ohdear:index:TagGroup` — full CRUD |
| `provider/cmd/pulumi-resource-ohdear/main.go` | Plugin entrypoint |
| `docs/_index.md` | Registry Overview tab |
| `docs/installation-configuration.md` | Registry Installation & Configuration tab |

## Develop

```sh
make build      # -> bin/pulumi-resource-ohdear
make install    # copy binary onto $GOPATH/bin
make schema     # -> schema.json
make sdk        # -> sdk/nodejs
```

`make install` then `pulumi package add ./bin/pulumi-resource-ohdear` in a stack to try it.

## Status

`Monitor`, `Tag` and `TagGroup` call the real API. `Monitor` covers the monitor
plus simple top-level fields (`checks`, `friendlyName`, `groupName`, `tags`,
`notes`, `description`, `realIpAddress`); the per-check `*_check_settings` trees
are not modelled yet. `teamId` and `type` are immutable (replace on change).
`Tag` has no update or delete endpoint upstream, so input changes replace it and
destroy only drops it from state. `Site` is still a stub. Resource tokens are
mapped to the `index` module.

Status page *updates* (`POST /api/status-page-updates`, the transient incident
messages) and the reusable update templates are not modelled — only the page
itself.

`CronCheck` is full CRUD against a monitor's `cron-checks` endpoints. `monitorId`
is immutable (replace on change) since the API can't move a check between
monitors. The imperative snooze/unsnooze and bulk `sync` endpoints are not
modelled. Read lists the parent monitor's checks, so importing one needs
`monitorId` supplied.

`RecurringMaintenancePeriod` is full CRUD for the recurring maintenance
*schedule* only — one-off maintenance periods are transient events, not desired
state. `monitorId` is immutable (accepted on create, not on update). The
`generate_now` / `regenerate_future_periods` / `delete_future_periods` backfill
knobs are left at their API defaults.

`NotificationDestination` picks its owner with `level`
(`team`/`monitor`/`tag`/`tagGroup`) + `ownerId`, both immutable. `destination`
is secret and kept from inputs on refresh (the API masks channel secrets); its
values are strings. `enabled` is reconciled through the enable/disable
endpoints, which never send a notification.
