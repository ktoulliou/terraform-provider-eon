# terraform-provider-eon

Terraform provider for **EyesOfNetwork** (EON) – manages Nagios monitoring objects via the EONAPI.

## Scope (v0.1 – minimal)

| Resource                     | EONAPI endpoints used                          |
|-----------------------------|-------------------------------------------------|
| `eon_host`                  | `createHost`, `getHost`, `deleteHost`           |
| `eon_command`               | `addCommand`, `getCommand`, `modifyCommand`, `deleteCommand` |
| `eon_contact`               | `createContact`, `getContact`, `modifyContact`, `deleteContact` |
| `eon_contact_group`         | `createContactGroup`, `getContactGroup`, `modifyContactGroup`, `deleteContactGroup` |
| `eon_export_configuration`  | `exportConfiguration`                           |

| Data Source    | Endpoint     |
|---------------|-------------|
| `eon_host`    | `getHost`   |
| `eon_command` | `getCommand`|

## Build & Install

```bash
# Build the binary
make build

# Install into local Terraform plugin directory
make install

# Verify
terraform init
```

## Provider Configuration

```hcl
provider "eon" {
  url      = "https://eon.example.com/eonapi"
  username = "admin"
  api_key  = "your-eon-api-key"
  insecure = true   # optional: skip TLS verify
}
```

All attributes can also be set via environment variables: `EON_URL`, `EON_USERNAME`, `EON_API_KEY`.

## Usage with existing Terraform variables

The main use case is feeding data from existing Terraform infrastructure into EON monitoring.
Use `for_each` with your variable maps:

```hcl
variable "servers" {
  type = map(object({
    ip       = string
    template = optional(string, "GENERIC_HOST")
  }))
}

resource "eon_host" "this" {
  for_each = var.servers
  name     = each.key
  ip       = each.value.ip
  template = each.value.template
}
```

### Workflow pattern

1. Create contacts & contact groups
2. Create check commands
3. Create hosts (referencing contacts/groups)
4. **Last**: trigger `eon_export_configuration` with `depends_on` to reload Nagios

```hcl
resource "eon_export_configuration" "apply" {
  job_name   = "terraform"
  depends_on = [eon_host.this, eon_command.this, eon_contact.this]
}
```

## Project structure

```
terraform-provider-eon/
├── main.go                              # Entry point
├── go.mod
├── Makefile
├── internal/
│   ├── client/
│   │   └── client.go                   # EONAPI HTTP client
│   └── provider/
│       ├── provider.go                 # Provider definition
│       ├── resource_host.go            # eon_host
│       ├── resource_command.go         # eon_command
│       ├── resource_contact.go         # eon_contact
│       ├── resource_contact_group.go   # eon_contact_group
│       ├── resource_export.go          # eon_export_configuration
│       └── datasources.go             # data sources
└── examples/
    └── main.tf                         # Full working example
```
