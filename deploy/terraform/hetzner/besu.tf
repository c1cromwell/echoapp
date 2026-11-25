# -------------------------------------------------------------------
# Echo's chain: four Hyperledger Besu validators (QBFT)
# -------------------------------------------------------------------
# Four is the smallest set that keeps running when one is lost. They are spread over the three
# sites of one network zone so they share the private network. With three sites, two validators
# share a site: losing that site stops the chain until one of them is back (nothing is lost).
# A fourth site in another zone would remove that, at the cost of a second private network.
#
# No chain port is public. The cloud firewall admits SSH only; peer, JSON-RPC and metrics traffic
# stays on the private network, where the host firewall (deploy/besu/ansible/site.yml) limits it
# to the other validators and the gateway subnet.
#
# Software and keys are installed by Ansible after the key ceremony, not here: a validator's key
# must never be in Terraform state.

variable "besu_enabled" {
  description = "Create the four chain validators"
  type        = bool
  default     = false
}

variable "besu_server_type" {
  description = "Server type for each validator. cpx22 (2 vCPU, 4 GB) is the floor: an idle validator uses about 1 GB. cpx32 (4 vCPU, 8 GB) if memory alerts fire."
  type        = string
  default     = "cpx22"
}

variable "besu_locations" {
  description = "Location of each of the four validators, in order. All must be in the eu-central network zone (nbg1, fsn1, hel1)."
  type        = list(string)
  default     = ["nbg1", "fsn1", "hel1", "nbg1"]
  validation {
    condition     = length(var.besu_locations) == 4
    error_message = "The chain is designed for exactly four validators."
  }
}

variable "besu_data_volume_gb" {
  description = "Size of each validator's chain data volume"
  type        = number
  default     = 40
}

variable "admin_cidrs" {
  description = "Addresses allowed to SSH to the validators. Narrow this to your own address."
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
}

resource "hcloud_network_subnet" "besu" {
  count        = var.besu_enabled ? 1 : 0
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.2.0/24"
}

resource "hcloud_firewall" "besu" {
  count = var.besu_enabled ? 1 : 0
  name  = "${local.name_prefix}-besu-fw"

  # SSH for the operator. Nothing else is reachable from the internet.
  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = var.admin_cidrs
  }
}

resource "hcloud_volume" "besu_data" {
  count    = var.besu_enabled ? 4 : 0
  name     = "${local.name_prefix}-besu-${count.index + 1}-data"
  size     = var.besu_data_volume_gb
  location = var.besu_locations[count.index]
  format   = "ext4"

  labels = {
    environment = var.environment
    role        = "besu-validator"
  }

  lifecycle {
    # The chain's data. Removing a validator from the configuration must not delete it.
    prevent_destroy = true
  }
}

resource "hcloud_server" "besu" {
  count        = var.besu_enabled ? 4 : 0
  name         = "${local.name_prefix}-besu-${count.index + 1}"
  server_type  = var.besu_server_type
  image        = "ubuntu-24.04"
  location     = var.besu_locations[count.index]
  ssh_keys     = [hcloud_ssh_key.default.id]
  firewall_ids = [hcloud_firewall.besu[0].id]

  network {
    network_id = hcloud_network.main.id
    ip         = "10.0.2.${11 + count.index}"
  }

  # Mount the data volume where the validator keeps the chain.
  user_data = <<-EOT
    #!/bin/bash
    set -euo pipefail
    mkdir -p /var/lib/echo-besu
    dev="/dev/disk/by-id/scsi-0HC_Volume_${hcloud_volume.besu_data[count.index].id}"
    until [ -e "$dev" ]; do sleep 2; done
    grep -q /var/lib/echo-besu /etc/fstab || echo "$dev /var/lib/echo-besu ext4 discard,nofail,defaults 0 0" >> /etc/fstab
    mount -a
  EOT

  labels = {
    environment = var.environment
    role        = "besu-validator"
  }

  depends_on = [hcloud_network_subnet.besu]
}

resource "hcloud_volume_attachment" "besu_data" {
  count     = var.besu_enabled ? 4 : 0
  volume_id = hcloud_volume.besu_data[count.index].id
  server_id = hcloud_server.besu[count.index].id
  automount = false
}

output "besu_public_ips" {
  description = "For SSH and the Ansible inventory (ansible_host)"
  value       = hcloud_server.besu[*].ipv4_address
}

output "besu_private_ips" {
  description = "For the Ansible inventory (private_ip) and the gateway's BESU_RPC_URLS"
  value       = [for i in range(var.besu_enabled ? 4 : 0) : "10.0.2.${11 + i}"]
}

output "besu_rpc_urls" {
  description = "Value for the gateway's BESU_RPC_URLS"
  value       = join(",", [for i in range(var.besu_enabled ? 4 : 0) : "http://10.0.2.${11 + i}:8545"])
}
