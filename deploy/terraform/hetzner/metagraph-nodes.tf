# -------------------------------------------------------------------
# Constellation metagraph node VMs (L0/L1 validators) + auto-generated Ansible inventory
# -------------------------------------------------------------------
# These nodes are provisioned here but CONFIGURED by hydra + Ansible (`hydra remote-deploy` /
# `remote-start`), which installs Java/deps and ships the metagraph JARs — so no cloud-init here.
# Disabled by default so backend-only applies never create them; enable with -var metagraph_enabled=true.
# The gl0 (Global L0) node is NOT provisioned here — Constellation recommends running it via `nodectl`
# (see docs/ZERO_TO_PRODUCTION.md §4). Populate euclid.json deploy.network.gl0_node before remote-start.

variable "metagraph_enabled" {
  description = "Provision the Constellation metagraph node VMs. Keep false for backend-only environments."
  type        = bool
  default     = false
}

variable "metagraph_node_count" {
  description = "Number of metagraph nodes. Constellation metagraphs run 3."
  type        = number
  default     = 3
}

variable "metagraph_node_type" {
  description = "Hetzner server type. Constellation spec ~8 vCPU / 16GB / 160GB. cpx41 = 8 vCPU(shared)/16GB/240GB (integrationnet-friendly); use a dedicated ccx33 (8 vCPU/32GB) for mainnet validators."
  type        = string
  default     = "cpx41"
}

variable "metagraph_ssh_user" {
  description = "Login user Ansible uses on the metagraph nodes."
  type        = string
  default     = "root"
}

variable "ansible_ssh_private_key_file" {
  description = "Private key Ansible uses to reach the metagraph nodes (must pair with var.ssh_public_key)."
  type        = string
  default     = "~/.ssh/id_ed25519"
}

variable "ansible_inventory_path" {
  description = "Where to render hosts.ansible.yml. Empty = deploy/terraform/hetzner/generated/hosts.ansible.yml; then copy into <euclid clone>/infra/ansible/remote/hosts.ansible.yml before `hydra remote-deploy`."
  type        = string
  default     = ""
}

locals {
  metagraph_inventory_path = var.ansible_inventory_path != "" ? var.ansible_inventory_path : "${path.module}/generated/hosts.ansible.yml"
}

# Metagraph nodes need public reachability for L0/L1 public + p2p (remote base ports 9100-9302).
resource "hcloud_firewall" "metagraph" {
  count = var.metagraph_enabled ? 1 : 0
  name  = "${local.name_prefix}-mg-fw"

  rule {
    direction  = "in"
    protocol   = "tcp"
    port       = "22"
    source_ips = ["0.0.0.0/0", "::/0"]
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "9100-9302"
    description = "Metagraph L0 / Currency L1 / Data L1 public + p2p + cli"
    source_ips  = ["0.0.0.0/0", "::/0"]
  }
}

resource "hcloud_server" "metagraph" {
  count        = var.metagraph_enabled ? var.metagraph_node_count : 0
  name         = "${local.name_prefix}-mg-${count.index + 1}"
  server_type  = var.metagraph_node_type
  image        = "ubuntu-24.04"
  location     = var.location
  ssh_keys     = [hcloud_ssh_key.default.id]
  firewall_ids = [hcloud_firewall.metagraph[0].id]

  labels = {
    environment = var.environment
    role        = "metagraph-node"
    index       = tostring(count.index + 1)
  }
}

# Render the Ansible inventory hydra remote-deploy/remote-start consume, from the live node IPs.
# Monitoring is co-located on node-1 by default (override the template for a dedicated host).
resource "local_file" "ansible_inventory" {
  count           = var.metagraph_enabled ? 1 : 0
  filename        = local.metagraph_inventory_path
  file_permission = "0644"
  content = templatefile("${path.module}/hosts.ansible.yml.tftpl", {
    node_ips      = hcloud_server.metagraph[*].ipv4_address
    ssh_user      = var.metagraph_ssh_user
    ssh_key       = var.ansible_ssh_private_key_file
    monitoring_ip = length(hcloud_server.metagraph) > 0 ? hcloud_server.metagraph[0].ipv4_address : ""
  })
}

output "metagraph_node_ips" {
  value = hcloud_server.metagraph[*].ipv4_address
}

output "metagraph_ansible_inventory" {
  value       = var.metagraph_enabled ? local.metagraph_inventory_path : ""
  description = "Copy this into <euclid clone>/infra/ansible/remote/hosts.ansible.yml"
}
