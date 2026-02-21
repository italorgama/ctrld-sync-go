# ctrld-hagezi-sync

Keeps your Control D profile(s) in sync with [Hagezi DNS blocklists](https://github.com/hagezi/dns-blocklists/tree/main/controld), automatically triggered whenever Hagezi releases an update.

Everything runs on GitHub Actions — no local setup, no `.env` file.

## Setup

### 1. Fork this repository

Click *Fork* on the top right of this page.

After forking, go to the *Actions* tab in your fork and enable workflows if prompted.

### 2. Get your Control D credentials

*API token*
1. Log in to your Control D account.
2. Go to *Preferences → API*.
3. Click *+* to create a new token and copy it.

*Profile ID*
1. Open the profile you want to sync.
2. Copy the ID from the URL:

```
https://controld.com/dashboard/profiles/abc123xyz/filters
                                        ^^^^^^^^^^^
```

### 3. Add repository secrets

Go to *Settings → Secrets and variables → Actions → New repository secret* and add:

| Secret    | Value                                                        |
|-----------|--------------------------------------------------------------|
| `TOKEN`   | Your Control D API token                                     |
| `PROFILE` | One or more profile IDs, comma-separated (e.g. `id1,id2`)   |

That's it. The workflows will run automatically from now on.

## How it works

| Workflow              | Trigger                         | What it does                                                  |
|-----------------------|---------------------------------|---------------------------------------------------------------|
| `check-release.yml`   | Every 2 hours                   | Checks for a new Hagezi release and triggers sync if detected |
| `sync.yml`            | Triggered by check, or manually | Builds the binary and runs the sync against your profile(s)   |

You can also trigger a manual sync anytime via *Actions → Sync → Run workflow*.

After each run, a summary with the number of folders and rules synced per profile is available under the *Summary* tab of the workflow run.

## Synced lists

Although this project ships pre-configured for Hagezi — chosen for the quality of its lists and the activity of the project — it supports any list in Control D's JSON folder format. Just add the raw URL to `lists.txt`.

Lists are configured in `lists.txt` — one URL per line. Lines starting with `#` are ignored.

The repository comes pre-configured with:

- Apple Private Relay Allow
- Native Trackers: Amazon, Apple, Huawei, LG WebOS, Microsoft, OPPO/Realme, Roku, Samsung, TikTok, Vivo, Xiaomi
- Ultimate Known Issues Allow
- Referral Allow
- Spam IDNs, Spam TLDs, Spam TLDs Allow
- Badware Hoster

To add or remove lists, edit `lists.txt`. Run `make list` to see all available Hagezi lists with their raw URLs ready to paste.

## License

MIT
