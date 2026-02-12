# ─── Provider configuration ─────────────────────────────────────
terraform {
  required_providers {
    eon = {
      source  = "eyesofnetwork/eon"
      version = "0.1.0"
    }
  }
}

provider "eon" {
  url      = var.eon_url       # https://192.168.1.100/eonapi
  username = var.eon_username   # admin
  api_key  = var.eon_api_key   # from /getApiKey
  insecure = true               # skip TLS verify (lab)
}

# ─── Variables (plug your existing Terraform variables here) ────
variable "eon_url" {
  type        = string
  description = "EONAPI base URL"
}

variable "eon_username" {
  type    = string
  default = "admin"
}

variable "eon_api_key" {
  type      = string
  sensitive = true
}

variable "servers" {
  description = "Map of servers to monitor"
  type = map(object({
    ip       = string
    alias    = optional(string, "")
    template = optional(string, "GENERIC_HOST")
  }))
  default = {
    "web-prod-01" = { ip = "10.0.1.10", alias = "Web frontend", template = "LINUX_SERVER" }
    "db-prod-01"  = { ip = "10.0.1.20", alias = "PostgreSQL primary" }
    "app-prod-01" = { ip = "10.0.1.30", alias = "Application server", template = "LINUX_SERVER" }
  }
}

variable "check_commands" {
  description = "Custom check commands"
  type = map(object({
    command_line = string
    description  = optional(string, "")
  }))
  default = {
    "check_app_health" = {
      command_line = "$USER1$/check_http -H $HOSTADDRESS$ -p 8080 -u /health -e 200"
      description  = "Application health endpoint"
    }
    "check_pg_replication" = {
      command_line = "$USER1$/check_pgsql -H $HOSTADDRESS$ -d replication -l nagios"
      description  = "PostgreSQL replication lag"
    }
  }
}

variable "contacts" {
  description = "Monitoring contacts"
  type = map(object({
    mail          = string
    alias         = optional(string, "")
    pager         = optional(string, "")
    contact_group = optional(string)
  }))
  default = {
    "oncall-ops" = {
      mail          = "oncall@example.com"
      alias         = "On-call Operations"
      contact_group = "admins"
    }
    "dev-lead" = {
      mail  = "dev-lead@example.com"
      alias = "Dev Team Lead"
    }
  }
}

# ─── Contact groups ─────────────────────────────────────────────
resource "eon_contact_group" "ops" {
  name        = "ops-team"
  description = "Operations team"
}

# ─── Contacts (from variable map) ──────────────────────────────
resource "eon_contact" "this" {
  for_each = var.contacts

  name          = each.key
  mail          = each.value.mail
  alias         = each.value.alias
  pager         = each.value.pager
  contact_group = each.value.contact_group
}

# ─── Check commands (from variable map) ─────────────────────────
resource "eon_command" "this" {
  for_each = var.check_commands

  name         = each.key
  command_line = each.value.command_line
  description  = each.value.description
}

# ─── Hosts (from variable map) ──────────────────────────────────
resource "eon_host" "this" {
  for_each = var.servers

  name          = each.key
  ip            = each.value.ip
  alias         = each.value.alias
  template      = each.value.template
  contact_group = eon_contact_group.ops.name
}

# ─── Export configuration (last step — reloads Nagios) ──────────
resource "eon_export_configuration" "apply" {
  job_name = "terraform-apply"

  depends_on = [
    eon_host.this,
    eon_command.this,
    eon_contact.this,
  ]
}

# ─── Data sources (read existing objects) ───────────────────────
data "eon_host" "web" {
  name = "web-prod-01"

  depends_on = [eon_host.this]
}

output "web_host_info" {
  value = data.eon_host.web.result_json
}
