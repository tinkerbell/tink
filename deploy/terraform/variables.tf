variable "packet_api_token" {
  description = "Packet user api token"
  type        = string
}

variable "project_id" {
  description = "Project ID"
  type        = string
}

variable "worker_count" {
  description = "Number of Workers"
  type        = number
  default     = 1
}
variable "facility" {
  description = "Packet facility to provision in"
  type        = string
  default     = "sjc1"
}

variable "device_type" {
  type        = string
  description = "Type of device to provision"
  default     = "c3.small.x86"
}

variable "ssh_user" {
  description = "Username that will be used to transfer file from your local environment to the provisioner"
  type        = string
  default     = "root"
}
