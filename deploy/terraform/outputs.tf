output "provisioner_dns_name" {
  value = "${split("-", packet_device.tink-provisioner.id)[0]}.packethost.net"
}

output "provisioner_ip" {
  value = packet_device.tink-provisioner.network[0].address
}

output "worker_mac_addr" {
  value = packet_device.tink-worker.ports[1].mac
}
