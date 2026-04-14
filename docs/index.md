---
page_title: "Provider: Websupport"
description: |-
  Manage DNS records on Websupport.sk via the public REST API.
---

# Websupport Provider

Manage DNS records on [Websupport.sk](https://websupport.sk).

## Example

```terraform
terraform {
  required_providers {
    websupport = {
      source  = "brainz-digital/websupport"
      version = "~> 0.1"
    }
  }
}

provider "websupport" {
  # api_key / api_secret are read from
  # WEBSUPPORT_API_KEY / WEBSUPPORT_SECRET when not set inline.
}
```

## Schema

### Optional

- `api_key` (String, Sensitive) — Websupport API key. Falls back to `WEBSUPPORT_API_KEY`.
- `api_secret` (String, Sensitive) — Websupport API secret. Falls back to `WEBSUPPORT_SECRET`.
- `base_url` (String) — Override the API base URL. Defaults to `https://rest.websupport.sk`.
