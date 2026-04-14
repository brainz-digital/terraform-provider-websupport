# Terraform Provider for Websupport.sk

Manage DNS records on [Websupport.sk](https://websupport.sk) via the public REST API.

Built with `terraform-plugin-framework`. Supports any record type (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, …) on a single `websupport_dns_record` resource.

## Usage

```hcl
terraform {
  required_providers {
    websupport = {
      source  = "brainz-digital/websupport"
      version = "~> 0.1"
    }
  }
}

provider "websupport" {
  # api_key / api_secret can be set inline,
  # or via WEBSUPPORT_API_KEY / WEBSUPPORT_SECRET env vars.
}

resource "websupport_dns_record" "apex" {
  zone    = "example.com"
  type    = "A"
  name    = "@"
  content = "203.0.113.10"
  ttl     = 600
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

## Authentication

API credentials are issued in the Websupport admin UI under **My account → API tokens**. Provide them via either:

- Provider block: `api_key` / `api_secret`
- Environment: `WEBSUPPORT_API_KEY` / `WEBSUPPORT_SECRET`

## Importing existing records

```bash
terraform import websupport_dns_record.apex example.com/12345
```

The import ID is `<zone>/<record_id>`. Find IDs by listing the zone:

```bash
curl -u "$WEBSUPPORT_API_KEY:$WEBSUPPORT_SECRET" \
  https://rest.websupport.sk/v1/user/self/zone/example.com/record | jq .
```

## Building from source

```bash
go build .
```

For local development, point Terraform at the build via `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/brainz-digital/websupport" = "/abs/path/to/repo"
  }
  direct {}
}
```

## License

[MPL-2.0](./LICENSE)
