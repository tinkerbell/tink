variable "auth_token" {
  description = "Packet api token."
  default     = "<API Auth token>"
}

variable "project_id" {
  description = "Id for your project"
  default = "<project id>"
}

variable "plan" {
  description = "Plan for the machine"
  default = "c3.small.x86"
}

variable "provisioner_ip" {
  type = string
  default     = "192.168.1.1"
}

variable "cidr" {
    type = number
    default = 29
}

variable "last_ip" {
  type = string
  default     = "192.168.1.7"
}

variable "netmask" {
  type = string
  default     = "255.255.255.248"
}

variable "private_registry_user" {
    type = string
    default = "admin"
}

variable "private_registry_pass" {
    type = string
    default = "admin123"
}
