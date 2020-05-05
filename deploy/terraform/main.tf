# Configure the Packet Provider.
provider "packet" {
  auth_token = var.packet_api_token
}

# Declare your project ID
locals {
  project_id = var.project_id
}

# Create a new VLAN in datacenter "ewr1"
resource "packet_vlan" "provisioning-vlan" {
  description = "provisioning-vlan"
  facility    = "sjc1"
  project_id  = local.project_id
}

# Create a device and add it to tf_project_1
resource "packet_device" "tf-provisioner" {
  hostname         = "tf-provisioner"
  plan             = "c3.small.x86"
  facilities       = ["sjc1"]
  operating_system = "ubuntu_18_04"
  billing_cycle    = "hourly"
  project_id       = local.project_id
  network_type     = "hybrid"
}

# Create a device and add it to tf_project_1
resource "packet_device" "tf-worker" {
  hostname         = "tf-worker"
  plan             = "c2.medium.x86"
  facilities       = ["sjc1"]
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