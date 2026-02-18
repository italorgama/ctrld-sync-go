.PHONY: list

list:
	@curl -s https://api.github.com/repos/hagezi/dns-blocklists/contents/controld \
		| jq -r '.[] | select(.name | endswith("-folder.json")) | "https://raw.githubusercontent.com/hagezi/dns-blocklists/main/controld/" + .name'
