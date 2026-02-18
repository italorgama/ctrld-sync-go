# ctrld-sync

Syncs [Hagezi DNS blocklists](https://github.com/hagezi/dns-blocklists) with your Control D profile(s). For each list, it deletes the existing folder and recreates it with the latest rules.

## Setup

Copy `.env.example` to `.env` and fill in your credentials:

```
TOKEN=your_control_d_api_token
PROFILE=profile_id_1,profile_id_2
```

- `TOKEN`: Control D API token
- `PROFILE`: One or more profile IDs separated by comma

## Running

```bash
go run main.go
```

Or build and run:

```bash
go build -o ctrld-sync main.go
./ctrld-sync
```

## Synced lists

From [hagezi/dns-blocklists](https://github.com/hagezi/dns-blocklists/tree/main/controld):

- Apple Private Relay Allow
- Native Trackers: Amazon, Apple, Huawei, LG WebOS, Microsoft, OPPO/Realme, Roku, Samsung, TikTok, Vivo, Xiaomi
- Ultimate Known Issues Allow
- Referral Allow
- Spam IDNs, Spam TLDs, Spam TLDs Allow
- Badware Hoster

To add or remove lists, edit `FolderURLs` in `main.go`.

## Automation

The included GitHub Actions workflows handle automatic syncing:

- `check-release.yml`: Runs every 2 hours and triggers a sync when a new Hagezi release is detected
- `sync.yml`: Runs the sync (triggered by `check-release.yml` or manually)

Required secrets: `TOKEN`, `PROFILE`

## License

MIT
