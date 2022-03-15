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


# Common SG 
resource "outscale_security_group" "node" {
  description = "Kubernetes node"

}

resource "outscale_security_group_rule" "node" {
  flow              = "Inbound"
  security_group_id = outscale_security_group.node.id
  # SSH
  rules {
    from_port_range = "22"
    to_port_range   = "22"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # Docker (Rancher)
  rules {
    from_port_range = "2376"
    to_port_range   = "2376"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # liveprobe
  rules {
    from_port_range = "9099"
    to_port_range   = "9099"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

}

# ETCD
resource "outscale_security_group" "etcd" {
  description = "Kubernetes etcd"
}

resource "outscale_security_group_rule" "etcd" {
  flow              = "Inbound"
  security_group_id = outscale_security_group.etcd.id

  # etcd
  rules {
    from_port_range = "2379"
    to_port_range   = "2380"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # Calico
  rules {
    ip_protocol = "-1"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

  rules {
    from_port_range = "179"
    to_port_range   = "179"
    ip_protocol     = "tcp"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

  rules {
    from_port_range = "4789"
    to_port_range   = "4789"
    ip_protocol     = "udp"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }
}

# Control plane SG
resource "outscale_security_group" "control-plane" {
  description = "Kubernetes control-planes"
}

resource "outscale_security_group_rule" "control-plane" {
  flow              = "Inbound"
  security_group_id = outscale_security_group.control-plane.id

  # kube-apiserver
  rules {
    from_port_range = "6443"
    to_port_range   = "6443"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # Calico
  rules {
    ip_protocol = "-1"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

  rules {
    from_port_range = "179"
    to_port_range   = "179"
    ip_protocol     = "tcp"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

  rules {
    from_port_range = "4789"
    to_port_range   = "4789"
    ip_protocol     = "udp"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }
}

# Worker
resource "outscale_security_group" "worker" {
  description = "Kubernetes workers"
}

resource "outscale_security_group_rule" "worker" {
  flow              = "Inbound"
  security_group_id = outscale_security_group.worker.security_group_id

  # Calico
  rules {
    ip_protocol = "-1"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

  rules {
    from_port_range = "179"
    to_port_range   = "179"
    ip_protocol     = "tcp"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

  rules {
    from_port_range = "4789"
    to_port_range   = "4789"
    ip_protocol     = "udp"
    security_groups_members {
      security_group_id = outscale_security_group.node.security_group_id
    }
  }

    # nginx HTTP
  rules {
    from_port_range = "80"
    to_port_range   = "80"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # nginx HTTPs
  rules {
    from_port_range = "443"
    to_port_range   = "443"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  # service node port range
  rules {
    from_port_range = "30000"
    to_port_range   = "32767"
    ip_protocol     = "tcp"
    ip_ranges       = ["0.0.0.0/0"]
  }

  rules {
    from_port_range = "30000"
    to_port_range   = "32767"
    ip_protocol     = "udp"
    ip_ranges       = ["0.0.0.0/0"]
  }
  
}

# Output
output "security_group_node" {
  value = outscale_security_group.node.id
  description = "Identification of the Security Group used by all nodes"
}

output "security_group_etcd" {
  value = outscale_security_group.etcd.id
  description = "Identification of the Security Group used by all etcd"
}

output "security_group_control_plane" {
  value = outscale_security_group.control-plane.id
  description = "Identification of the Security Group used by all control-planes"
}

output "security_group_worker" {
  value = outscale_security_group.worker.id
  description = "Identification of the Security Group used by all workers"
}
