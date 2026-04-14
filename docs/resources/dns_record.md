---
page_title: "websupport_dns_record Resource - websupport"
description: |-
  A DNS record managed via the Websupport REST API.
---

# websupport_dns_record (Resource)

Manages a single DNS record in a Websupport-hosted zone. Supports any record type the API accepts (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, …).

Changing `zone`, `type`, or `name` forces replacement; `content`, `ttl`, `priority`, and `note` update in place.

## Example Usage

```terraform
resource "websupport_dns_record" "apex" {
  zone    = "example.com"
  type    = "A"
  name    = "@"
  content = "203.0.113.10"
  ttl     = 600
}

resource "websupport_dns_record" "wildcard" {
  zone    = "example.com"
  type    = "A"
  name    = "*"
  content = "203.0.113.10"
}

resource "websupport_dns_record" "mx" {
  zone     = "example.com"
  type     = "MX"
  name     = "@"
  content  = "mail.example.com"
  priority = 10
  ttl      = 3600
}
```

## Schema

### Required

- `zone` (String) — Zone (domain) the record belongs to. Forces replacement.
- `type` (String) — Record type. Forces replacement.
- `name` (String) — Record name. Use `@` for the apex and `*` for wildcard. Forces replacement.
- `content` (String) — Record value (IP, hostname, text, etc).

### Optional

- `ttl` (Number) — TTL in seconds. Defaults to `600`.
- `priority` (Number) — Priority for `MX` / `SRV` records.
- `note` (String) — Free-form note stored alongside the record.

### Read-Only

- `id` (String) — Numeric record ID assigned by Websupport.

## Import

```bash
terraform import websupport_dns_record.apex example.com/12345
```

The import ID is `<zone>/<record_id>`.
