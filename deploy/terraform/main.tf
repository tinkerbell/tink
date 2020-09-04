# Configure the Packet Provider.
terraform {
  required_providers {
    packet = {
      source  = "packethost/packet"
      version = "~> 3.0.1"
    }
    null = {
      source = "hashicorp/null"
    }
  }
}

provider "packet" {
  auth_token = var.packet_api_token
}

# Create a new VLAN in datacenter "ewr1"
resource "packet_vlan" "provisioning_vlan" {
  description = "provisioning_vlan"
  facility    = var.facility
  project_id  = var.project_id
}

# Create a device and add it to tf_project_1
resource "packet_device" "tink_provisioner" {
  hostname         = "tink-provisioner"
  plan             = var.device_type
  facilities       = [var.facility]
  operating_system = "ubuntu_18_04"
  billing_cycle    = "hourly"
  project_id       = var.project_id
  user_data        = file("install_package.sh")
}

resource "null_resource" "tink_directory" {
  connection {
    type = "ssh"
    user = var.ssh_user
    host = packet_device.tink_provisioner.network[0].address
  }

  provisioner "file" {
    source      = "./../../../tink"
    destination = "/root/"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x tink/generate-envrc.sh tink/setup.sh tink/deploy/tls/*.sh"
    ]
  }
}

resource "packet_device_network_type" "tink_provisioner_network_type" {
  device_id = packet_device.tink_provisioner.id
  type      = "hybrid"
}

# Create a device and add it to tf_project_1
resource "packet_device" "tink_worker" {
  count = var.worker_count

  hostname         = "tink-worker-${count.index}"
  plan             = var.device_type
  facilities       = [var.facility]
  operating_system = "custom_ipxe"
  ipxe_script_url  = "https://boot.netboot.xyz"
  always_pxe       = "true"
  billing_cycle    = "hourly"
  project_id       = var.project_id
}

resource "packet_device_network_type" "tink_worker_network_type" {
  count = var.worker_count

  device_id = packet_device.tink_worker[count.index].id
  type      = "layer2-individual"
}

# Attach VLAN to provisioner
resource "packet_port_vlan_attachment" "provisioner" {
  depends_on = [packet_device_network_type.tink_provisioner_network_type]
  device_id  = packet_device.tink_provisioner.id
  port_name  = "eth1"
  vlan_vnid  = packet_vlan.provisioning_vlan.vxlan
}

# Attach VLAN to worker
resource "packet_port_vlan_attachment" "worker" {
  count      = var.worker_count
  depends_on = [packet_device_network_type.tink_worker_network_type]

  device_id = packet_device.tink_worker[count.index].id
  port_name = "eth0"
  vlan_vnid = packet_vlan.provisioning_vlan.vxlan
}

data "template_file" "worker_hardware_data" {
  count    = var.worker_count
  template = file("${path.module}/hardware_data.tpl")
  vars = {
    id            = packet_device.tink_worker[count.index].id
    facility_code = packet_device.tink_worker[count.index].deployed_facility
    plan_slug     = packet_device.tink_worker[count.index].plan
    address       = "192.168.1.${count.index + 5}"
    mac           = packet_device.tink_worker[count.index].ports[1].mac
  }
}

resource "null_resource" "hardware_data" {
  count = var.worker_count
  connection {
    type = "ssh"
    user = var.ssh_user
    host = packet_device.tink_provisioner.network[0].address
  }

  provisioner "file" {
    content     = data.template_file.worker_hardware_data[count.index].rendered
    destination = "/root/tink/deploy/hardware-data-${count.index}.json"
  }
}
