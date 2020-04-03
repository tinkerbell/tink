# Configure the Packet Provider.
provider "packet" {
  auth_token = var.auth_token
}

# Declare your project ID
locals {
  project_id = var.project_id
}

# Create a new VLAN in datacenter "ewr1"
resource "packet_vlan" "provisioning-vlan" {
  description = "provisioning-vlan"
  facility    = "dfw2"
  project_id  = local.project_id
}

# Create a device and add it to tf_project_1
resource "packet_device" "tf-provisioner" {
  hostname         = "tf-provisioner"
  plan             = var.plan
  facilities       = ["dfw2"]
  operating_system = "ubuntu_18_04"
  billing_cycle    = "hourly"
  project_id       = local.project_id
  network_type     = "hybrid"
  
  connection {
    user = "root"
    password = packet_device.tf-provisioner.root_password
    host = packet_device.tf-provisioner.network[0].address
  }
  
  provisioner "file" {
    source      = "../cmd/tinkerbell/tink-cli"
    destination = "/usr/local/bin/tink"
  }

  provisioner "remote-exec" {
    inline = [
        "echo \"HOST_IP=${var.provisioner_ip}\" >> /etc/environment",
        "echo \"IP_CIDR=${var.cidr}\" >> /etc/environment",
        "echo \"BROAD_IP=${var.last_ip}\" >> /etc/environment",
        "echo \"NETMASK=${var.netmask}\" >> /etc/environment",
        "echo \"TINKERBELL_REGISTRY_USER=${var.private_registry_user}\" >> /etc/environment",
        "echo \"TINKERBELL_REGISTRY_PASS=${var.private_registry_pass}\" >> /etc/environment",
        "echo \"TINKERBELL_GRPC_AUTHORITY=\"127.0.0.1:42113\"\" >> /etc/environment",
        "echo \"TINKERBELL_CERT_URL=\"http://127.0.0.1:42114/cert\"\" >> /etc/environment",
        "cat /etc/environment"
        ]
  }
  provisioner "remote-exec" {
    script = "setup.sh"
  }
}

# Create a device and add it to tf_project_1
resource "packet_device" "tf-worker" {
  hostname         = "tf-worker"
  plan             = var.plan
  facilities       = ["dfw2"]
  operating_system = "custom_ipxe"
  ipxe_script_url  = "https://boot.netboot.xyz"
  always_pxe       = "true"
  billing_cycle    = "hourly"
  project_id       = local.project_id
  network_type     = "layer2-individual"
}

# Attach VLAN to provisioner
resource "packet_port_vlan_attachment" "provisioner" {
  device_id = packet_device.tf-provisioner.id
  port_name = "eth1"
  vlan_vnid = packet_vlan.provisioning-vlan.vxlan
}

# Attach VLAN to worker
resource "packet_port_vlan_attachment" "worker" {
  device_id = packet_device.tf-worker.id
  port_name = "eth0"
  vlan_vnid = packet_vlan.provisioning-vlan.vxlan
}

output "provisioner_ip" {
  value = "${packet_device.tf-provisioner.network[0].address}"
}

output "worker_mac_addr" {
  value = "${packet_device.tf-worker.ports[1].mac}"
}
