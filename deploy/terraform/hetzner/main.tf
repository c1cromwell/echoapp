terraform {
  required_version = ">= 1.5.0"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.45"
    }
  }
}

provider "hcloud" {
  token = var.hcloud_token
}

# -------------------------------------------------------------------
# Variables
# -------------------------------------------------------------------

variable "hcloud_token" {
  description = "Hetzner Cloud API token"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name: dev, staging, or prod"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "environment must be dev, staging, or prod"
  }
}

variable "ssh_public_key" {
  description = "SSH public key for server access"
  type        = string
}

variable "location" {
  description = "Hetzner datacenter location"
  type        = string
  default     = "nbg1" # Nuremberg (also: fsn1, hel1, ash)
}

locals {
  name_prefix = "echo-${var.environment}"

  server_config = {
    dev     = { type = "cpx21",  k3s_nodes = 1, db_type = "cpx11"  }
    staging = { type = "cpx31",  k3s_nodes = 2, db_type = "cpx21"  }
    prod    = { type = "cpx41",  k3s_nodes = 3, db_type = "cpx31"  }
  }

  cfg = local.server_config[var.environment]
}

# -------------------------------------------------------------------
# SSH Key
# -------------------------------------------------------------------

resource "hcloud_ssh_key" "default" {
  name       = "${local.name_prefix}-key"
  public_key = var.ssh_public_key
}

# -------------------------------------------------------------------
# Network
# -------------------------------------------------------------------

resource "hcloud_network" "main" {
  name     = "${local.name_prefix}-network"
  ip_range = "10.0.0.0/16"
}

resource "hcloud_network_subnet" "k3s" {
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.1.0/24"
}

# -------------------------------------------------------------------
# Firewall
# -------------------------------------------------------------------

resource "hcloud_firewall" "k3s" {
  name = "${local.name_prefix}-fw"

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "6443"
    description = "k3s API server"
    source_ips = ["0.0.0.0/0", "::/0"]
  }
}

# -------------------------------------------------------------------
# k3s Server (control plane)
# -------------------------------------------------------------------

resource "hcloud_server" "k3s_server" {
  name        = "${local.name_prefix}-k3s-server"
  server_type = local.cfg.type
  image       = "ubuntu-24.04"
  location    = var.location
  ssh_keys    = [hcloud_ssh_key.default.id]
  firewall_ids = [hcloud_firewall.k3s.id]

  network {
    network_id = hcloud_network.main.id
    ip         = "10.0.1.10"
  }

  user_data = <<-EOF
    #!/bin/bash
    set -euo pipefail

    # Install k3s (server / control plane)
    curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server \
      --disable traefik \
      --tls-san ${local.name_prefix}-k3s-server \
      --node-name ${local.name_prefix}-k3s-server \
      --flannel-iface ens10" sh -

    # Wait for k3s to be ready
    until kubectl get nodes &>/dev/null; do sleep 2; done

    # Install cert-manager for TLS
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml

    # Store k3s token for agent nodes
    cat /var/lib/rancher/k3s/server/node-token > /root/k3s-token
  EOF

  labels = {
    environment = var.environment
    role        = "k3s-server"
  }
}

# -------------------------------------------------------------------
# k3s Agent Nodes (workers)
# -------------------------------------------------------------------

resource "hcloud_server" "k3s_agent" {
  count       = local.cfg.k3s_nodes > 1 ? local.cfg.k3s_nodes - 1 : 0
  name        = "${local.name_prefix}-k3s-agent-${count.index + 1}"
  server_type = local.cfg.type
  image       = "ubuntu-24.04"
  location    = var.location
  ssh_keys    = [hcloud_ssh_key.default.id]
  firewall_ids = [hcloud_firewall.k3s.id]

  network {
    network_id = hcloud_network.main.id
    ip         = "10.0.1.${11 + count.index}"
  }

  user_data = <<-EOF
    #!/bin/bash
    set -euo pipefail

    # Wait for server to be ready and retrieve token
    until curl -sf http://10.0.1.10:6443/healthz &>/dev/null; do sleep 5; done

    # Install k3s agent
    curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="agent \
      --server https://10.0.1.10:6443 \
      --node-name ${local.name_prefix}-k3s-agent-${count.index + 1} \
      --flannel-iface ens10" \
      K3S_TOKEN_FILE=/dev/stdin sh - <<< "$(ssh -o StrictHostKeyChecking=no root@10.0.1.10 cat /root/k3s-token)"
  EOF

  labels = {
    environment = var.environment
    role        = "k3s-agent"
  }

  depends_on = [hcloud_server.k3s_server]
}

# -------------------------------------------------------------------
# Database Server (PostgreSQL + Redis)
# -------------------------------------------------------------------

resource "hcloud_server" "db" {
  name        = "${local.name_prefix}-db"
  server_type = local.cfg.db_type
  image       = "ubuntu-24.04"
  location    = var.location
  ssh_keys    = [hcloud_ssh_key.default.id]

  network {
    network_id = hcloud_network.main.id
    ip         = "10.0.1.50"
  }

  user_data = <<-EOF
    #!/bin/bash
    set -euo pipefail

    # Install PostgreSQL 16
    apt-get update
    apt-get install -y postgresql-16 redis-server

    # Configure PostgreSQL to listen on private network
    echo "listen_addresses = '10.0.1.50'" >> /etc/postgresql/16/main/postgresql.conf
    echo "host echoapp echoapp 10.0.1.0/24 scram-sha-256" >> /etc/postgresql/16/main/pg_hba.conf

    sudo -u postgres psql -c "CREATE USER echoapp WITH PASSWORD 'CHANGE_ME';"
    sudo -u postgres psql -c "CREATE DATABASE echoapp OWNER echoapp;"

    systemctl restart postgresql

    # Configure Redis to listen on private network
    sed -i 's/^bind .*/bind 10.0.1.50/' /etc/redis/redis.conf
    sed -i 's/^# maxmemory .*/maxmemory 256mb/' /etc/redis/redis.conf
    sed -i 's/^# maxmemory-policy .*/maxmemory-policy allkeys-lru/' /etc/redis/redis.conf
    systemctl restart redis-server
  EOF

  labels = {
    environment = var.environment
    role        = "database"
  }
}

# -------------------------------------------------------------------
# Load Balancer
# -------------------------------------------------------------------

resource "hcloud_load_balancer" "main" {
  name               = "${local.name_prefix}-lb"
  load_balancer_type = "lb11"
  location           = var.location

  labels = {
    environment = var.environment
  }
}

resource "hcloud_load_balancer_network" "main" {
  load_balancer_id = hcloud_load_balancer.main.id
  network_id       = hcloud_network.main.id
  ip               = "10.0.1.100"
}

resource "hcloud_load_balancer_target" "k3s" {
  load_balancer_id = hcloud_load_balancer.main.id
  type             = "server"
  server_id        = hcloud_server.k3s_server.id
}

resource "hcloud_load_balancer_service" "https" {
  load_balancer_id = hcloud_load_balancer.main.id
  protocol         = "https"
  listen_port      = 443
  destination_port = 80

  health_check {
    protocol = "http"
    port     = 80
    interval = 15
    timeout  = 10
    retries  = 3

    http {
      path         = "/health"
      status_codes = ["200"]
    }
  }

  http {
    redirect_http = true
  }
}

# -------------------------------------------------------------------
# Volumes (persistent storage)
# -------------------------------------------------------------------

resource "hcloud_volume" "db_data" {
  name      = "${local.name_prefix}-db-data"
  size      = var.environment == "prod" ? 100 : 20
  server_id = hcloud_server.db.id
  automount = true
  format    = "ext4"
  location  = var.location
}

# -------------------------------------------------------------------
# Outputs
# -------------------------------------------------------------------

output "k3s_server_ip" {
  value = hcloud_server.k3s_server.ipv4_address
}

output "db_private_ip" {
  value = "10.0.1.50"
}

output "load_balancer_ip" {
  value = hcloud_load_balancer.main.ipv4
}

output "k3s_agent_ips" {
  value = hcloud_server.k3s_agent[*].ipv4_address
}
