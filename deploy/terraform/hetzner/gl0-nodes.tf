# -------------------------------------------------------------------
# Echo's own Global L0 hosts and the private chain network (docs/OWN_L0_PLAN.md stage 2)
# -------------------------------------------------------------------
# Echo runs its own Global L0 instead of joining Constellation's. These hosts are provisioned here
# and configured by deploy/ansible (not hydra: Euclid has no remote Global L0 playbook).
#
# Network rule: chain ports are never opened on a public address. Every chain node sits on the
# private network (10.0.2.0/24) and peers over it, and the gateway on k3s (10.0.1.0/24) reaches
# the nodes the same way. Hetzner firewalls filter only the public interface, so the firewall below
# admits SSH and nothing else. Ansible adds a host firewall that limits the chain ports to
# 10.0.0.0/16 as a second layer.
#
# Disabled by default. Enable with -var chain_enabled=true (and metagraph_enabled=true for the
# metagraph hosts in metagraph-nodes.tf).

variable "chain_enabled" {
  description = "Provision Echo's own Global L0 hosts and attach the metagraph hosts to the private chain network."
  type        = bool
  default     = false
}

variable "gl0_node_count" {
  description = "Number of Global L0 nodes. Three tolerates the loss of any one."
  type        = number
  default     = 3
}

variable "gl0_colocate" {
  description = "Run the Global L0 nodes on the metagraph hosts instead of on their own. Saves gl0_node_count servers; a lost host then takes a Global L0 node and a metagraph node down together, and each host needs memory for both (about 11 GB of JVM heap)."
  type        = bool
  default     = false
}

# No default: Hetzner's prices and the set of orderable types changed several times in 2026.
# Choose from what the Cloud Console offers your account on the day. A Global L0 node needs
# about 4 vCPU and 8 GB.
variable "gl0_node_type" {
  description = "Hetzner server type for a Global L0 node (ignored when gl0_colocate = true)."
  type        = string
  default     = ""
}

variable "gl0_locations" {
  description = "One location per Global L0 node, used in order and repeated if shorter. All must be in the eu-central network zone (nbg1, fsn1, hel1) so they share the private network. Empty = var.location for all."
  type        = list(string)
  default     = []
}

variable "admin_cidrs" {
  description = "Source addresses allowed to SSH to chain hosts. Narrow this to the operator's addresses for production."
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
}

locals {
  gl0_count     = var.chain_enabled && !var.gl0_colocate ? var.gl0_node_count : 0
  gl0_locations = length(var.gl0_locations) > 0 ? var.gl0_locations : [var.location]
  # Fixed private addresses: Global L0 at .11-.19, metagraph hosts at .21-.29.
  gl0_private_ips       = [for i in range(local.gl0_count) : "10.0.2.${11 + i}"]
  metagraph_private_ips = [for i in range(var.metagraph_enabled ? var.metagraph_node_count : 0) : "10.0.2.${21 + i}"]
  chain_inventory_path  = "${path.module}/generated/chain.inventory.yml"
}

resource "hcloud_network_subnet" "chain" {
  count        = var.chain_enabled ? 1 : 0
  network_id   = hcloud_network.main.id
  type         = "cloud"
  network_zone = "eu-central"
  ip_range     = "10.0.2.0/24"
}

resource "hcloud_firewall" "chain" {
  count = var.chain_enabled ? 1 : 0
  name  = "${local.name_prefix}-chain-fw"

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "22"
    description = "SSH for operators. No chain port is public."
    source_ips  = var.admin_cidrs
  }
}

resource "hcloud_server" "gl0" {
  count        = local.gl0_count
  name         = "${local.name_prefix}-gl0-${count.index + 1}"
  server_type  = var.gl0_node_type
  image        = "ubuntu-24.04"
  location     = local.gl0_locations[count.index % length(local.gl0_locations)]
  ssh_keys     = [hcloud_ssh_key.default.id]
  firewall_ids = [hcloud_firewall.chain[0].id]

  network {
    network_id = hcloud_network.main.id
    ip         = local.gl0_private_ips[count.index]
  }

  labels = {
    environment = var.environment
    role        = "global-l0"
    index       = tostring(count.index + 1)
  }

  depends_on = [hcloud_network_subnet.chain]

  lifecycle {
    precondition {
      condition     = var.gl0_node_type != ""
      error_message = "Set gl0_node_type to a server type the Hetzner Cloud Console currently offers, or set gl0_colocate = true."
    }
  }
}

# Inventory for deploy/ansible. ansible_host is the public address (SSH); chain_ip is the private
# address every node advertises and peers on.
resource "local_file" "chain_inventory" {
  count           = var.chain_enabled ? 1 : 0
  filename        = local.chain_inventory_path
  file_permission = "0644"
  content = templatefile("${path.module}/chain.inventory.yml.tftpl", {
    mg = [for i, s in hcloud_server.metagraph : { ssh = s.ipv4_address, ip = local.metagraph_private_ips[i] }]
    # Co-located: the Global L0 entries point at the first gl0_node_count metagraph hosts.
    gl0 = (var.gl0_colocate
      ? [for i, s in slice(hcloud_server.metagraph, 0, min(var.gl0_node_count, length(hcloud_server.metagraph))) : { ssh = s.ipv4_address, ip = local.metagraph_private_ips[i] }]
      : [for i, s in hcloud_server.gl0 : { ssh = s.ipv4_address, ip = local.gl0_private_ips[i] }]
    )
    ssh_user = var.metagraph_ssh_user
    ssh_key  = var.ansible_ssh_private_key_file
    network  = var.environment
  })
}

output "gl0_node_ips" {
  value = hcloud_server.gl0[*].ipv4_address
}

output "gl0_private_ips" {
  value = var.gl0_colocate ? slice(local.metagraph_private_ips, 0, min(var.gl0_node_count, length(local.metagraph_private_ips))) : local.gl0_private_ips
}

output "metagraph_private_ips" {
  value = local.metagraph_private_ips
}

output "chain_ansible_inventory" {
  value       = var.chain_enabled ? local.chain_inventory_path : ""
  description = "Pass to deploy/ansible with -i"
}
