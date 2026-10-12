# Q-Panel customization (based on Sun-Panel v1.3.0)

## Summary

Three features replaced the original global LAN/WAN switch:

1. **Per-group multi network states** – each icon group owns a list of network states (公网 / 内网 / 异地组网 / custom) and an active state.
2. **Item + group form fields** – items keep a full public `url` and a `port`; groups edit state name + host/IP.
3. **Cross-group drag sort** – in sort mode, icons can be dragged between all groups; drop updates `itemIconGroupId` + `sort`.

## Data model

### `ItemIconGroup.networkConfig` (JSON text column)

```json
{
  "networkStates": [
    { "id": "wan", "name": "公网", "type": "wan", "host": "" },
    { "id": "lan", "name": "内网", "type": "lan", "host": "192.168.1.10" },
    { "id": "mesh1", "name": "异地组网", "type": "custom", "host": "10.8.0.2" }
  ],
  "activeNetworkStateId": "wan"
}
```

- `type`: `wan` | `lan` | `custom`
- New groups default to wan + lan.
- Existing groups with empty config are backfilled on `getList`.

### `ItemIcon.port` (int)

- New column; `lanUrl` column kept for DB compatibility but unused in the new UI.
- Public address remains full `url` (`http(s)://...`).

## Jump URL resolution

```
if activeState.type === 'wan' → item.url
else → take protocol + path from item.url, replace hostname with state.host,
       replace port with item.port (if set); fall back to item.url if host empty
```

Implemented in `src/utils/network/index.ts` (`buildJumpUrl` / `buildUrlForState`).

Right-click menu offers opening with other states of the same group, plus “open public URL”.

## Cross-group sort

- Home page uses a **global sort mode** (any group’s drag icon toggles all groups).
- `VueDraggable` shares `group: panel-icons` so items move across lists.
- `POST /panel/itemIcon/saveSort` accepts:

```json
{
  "sortItems": [
    { "id": 1, "sort": 1, "itemIconGroupId": 2 },
    { "id": 3, "sort": 2, "itemIconGroupId": 2 }
  ]
}
```

After a move, clicks use the **destination group’s** active network state + the item’s port/url.

## Migration notes

- GORM `AutoMigrate` adds `network_config` on `item_icon_group` and `port` on `item_icon`.
- No manual SQL required for SQLite/MySQL when the service starts normally.
- Old exports with `lanUrl` still import; `port` defaults to 0. Prefer re-exporting after upgrade to carry `networkConfig` + `port`.
- Global floating LAN/WAN buttons are removed. Style setting `netModeChangeButtonShow` remains but has no effect on the home floating bar.

## Key files

| Area | Path |
|------|------|
| Group model | `service/models/itemIconGroup.go` |
| Network JSON type | `service/models/datatype/networkConfig.go` |
| Item model | `service/models/itemIcon.go` |
| Group API | `service/api/api_v1/panel/ItemIconGroup.go` |
| Item sort API | `service/api/api_v1/panel/itemIcon.go` |
| URL helper | `src/utils/network/index.ts` |
| Home | `src/views/home/index.vue` |
| Edit item | `src/views/home/components/EditItem/index.vue` |
| Group manage | `src/components/apps/ItemGroupManage/index.vue` |
