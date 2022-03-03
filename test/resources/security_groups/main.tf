terraform {
  required_providers {
    outscale = {
      source  = "outscale-dev/outscale"
    }
  }
}

provider "outscale" {
  access_key_id = var.access_key_id
  secret_key_id = var.secret_key_id
  region        = var.region
}

variable "access_key_id" {}
variable "secret_key_id" {}
variable "region" {}

resource "outscale_security_group" "test-docker" {
  description = "test-docker"
}

resource "outscale_security_group_rule" "test-docker" {
  flow              = "Inbound"
  security_group_id = outscale_security_group.test-docker.security_group_id
  rules {
    from_port_range = "22"
    to_port_range   = "22"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # Docker
  rules {
    from_port_range = "2376"
    to_port_range   = "2376"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

}