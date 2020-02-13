variable "provisioner_ip" {
  type = string
  default     = "192.168.2.1"
}

variable "cidr" {
    type = number
    default = 28
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

variable "rover_registry_user" {
    type = string
    default = "roveradmin"
}

variable "rover_registry_pass" {
    type = string
    default = "rover123"
}
