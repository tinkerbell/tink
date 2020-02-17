variable "auth_token" {
  description = "Packet api token."
  default     = "<Authentication token for packet>"
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

// This variable will be removed once the quay repository will be opensourced
variable "quay_user" {
    type = string
    default = "quay_username"
}

// This variable will be removed once the quay repository will be opensourced 
variable "quay_pass" {
    type = string
    default = "quay_password"
}

// This variable will be removed once the quay repository will be opensourced
variable "git_user" {
    type = string
    default = "github_username"
}

// This variable will be removed once the quay repository will be opensourced
variable "git_pass" {
    type = string
    default = "github_password"
}

variable "private_registry_user" {
    type = string
    default = "admin"
}

variable "private_registry_pass" {
    type = string
    default = "admin123"
}
