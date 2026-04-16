variable "ENVIRONMENT_ID" {
  description = "The environment ID."
  type        = string
}

resource "dynatrace_environment_ip_allowlist" "test" {
  environment_id         = var.ENVIRONMENT_ID
  enabled                = true
  allow_webhook_override = false

  allowlist {
    name     = "office"
    ip_range = "10.0.0.0/8"
  }
}
