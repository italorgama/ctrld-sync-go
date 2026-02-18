# ctrld-hagezi-sync

Keeps your Control D profile(s) in sync with [Hagezi DNS blocklists](https://github.com/hagezi/dns-blocklists/tree/main/controld), automatically triggered whenever Hagezi releases an update.

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

Lists are configured in `lists.txt` — one URL per line. Lines starting with `#` are ignored.

The repository comes pre-configured with:

- Apple Private Relay Allow
- Native Trackers: Amazon, Apple, Huawei, LG WebOS, Microsoft, OPPO/Realme, Roku, Samsung, TikTok, Vivo, Xiaomi
- Ultimate Known Issues Allow
- Referral Allow
- Spam IDNs, Spam TLDs, Spam TLDs Allow
- Badware Hoster

To add or remove lists, edit `lists.txt`. Run `make list` to see all available Hagezi lists with their raw URLs ready to paste.

## Automation

Fork this repository to use the included GitHub Actions workflows:

- `check-release.yml`: Runs every 2 hours and triggers a sync when a new Hagezi release is detected
- `sync.yml`: Runs the sync (triggered by `check-release.yml` or manually)

After forking, add `TOKEN` and `PROFILE` to your repository secrets under **Settings → Secrets and variables → Actions**.

## License

MIT
