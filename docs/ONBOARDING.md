# Toni Onboarding

On first run, `toni` opens a setup TUI that helps users configure YELP autocomplete.

## What it asks

1. Enable or disable YELP API autocomplete
2. If enabled, paste a YELP API key (or skip)

## Setup help link

During onboarding, users are guided to:

- https://github.com/mihirchanduka/toni

## Secure key storage

When a key is provided, `toni` stores it in:

- `~/.toni/yelp_api_key`

Permissions:

- Directory: `~/.toni` uses `0700`
- Key file: `~/.toni/yelp_api_key` uses `0600`

The onboarding status (non-secret) is stored in:

- `~/.toni/onboarding.json`
