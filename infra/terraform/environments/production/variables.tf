variable "region" { type = string }
variable "oci_profile" {
  type    = string
  default = "DEFAULT"
}
variable "tenancy_id" { type = string }
variable "compartment_id" { type = string }
variable "availability_domain_index" {
  type    = number
  default = 0
  validation {
    condition     = var.availability_domain_index >= 0 && floor(var.availability_domain_index) == var.availability_domain_index
    error_message = "availability_domain_index must be a nonnegative integer."
  }
}
variable "admin_ipv4_cidr" {
  type = string
  validation {
    condition     = can(cidrhost(var.admin_ipv4_cidr, 0)) && endswith(var.admin_ipv4_cidr, "/32")
    error_message = "admin_ipv4_cidr must be one explicit IPv4 /32."
  }
}
variable "cloudflare_ipv4_cidrs" {
  type = set(string)
  validation {
    condition     = length(var.cloudflare_ipv4_cidrs) > 0 && alltrue([for cidr in var.cloudflare_ipv4_cidrs : can(cidrhost(cidr, 0)) && !strcontains(cidr, ":")])
    error_message = "cloudflare_ipv4_cidrs must contain reviewed IPv4 CIDRs."
  }
}
variable "ssh_public_key_path" { type = string }
variable "compute_shape" {
  type    = string
  default = "VM.Standard.E2.1.Micro"
  validation {
    condition     = var.compute_shape == "VM.Standard.E2.1.Micro"
    error_message = "Only VM.Standard.E2.1.Micro is permitted by the Always Free design."
  }
}
variable "compute_image_id" {
  type        = string
  default     = null
  nullable    = true
  description = "Optional pinned Oracle Linux image OCID. Null selects the newest matching Oracle Linux 9 image."
}
variable "snowflake_node_id" {
  type    = number
  default = 0
  validation {
    condition     = var.snowflake_node_id >= 0 && var.snowflake_node_id <= 1023 && floor(var.snowflake_node_id) == var.snowflake_node_id
    error_message = "snowflake_node_id must be an integer from 0 through 1023."
  }
}
variable "mysql_shape" {
  type    = string
  default = "MySQL.Free"
  validation {
    condition     = var.mysql_shape == "MySQL.Free"
    error_message = "Only MySQL.Free is permitted by the Always Free design."
  }
}
variable "mysql_admin_username" {
  type    = string
  default = "ostos_admin"
}
variable "mysql_admin_password" {
  type        = string
  sensitive   = true
  description = "Supply only through TF_VAR_mysql_admin_password or the CI secret store."
}
variable "vcn_cidr" {
  type    = string
  default = "10.0.0.0/16"
}
variable "public_subnet_cidr" {
  type    = string
  default = "10.0.1.0/24"
}
variable "database_subnet_cidr" {
  type    = string
  default = "10.0.2.0/24"
}
variable "compute_private_ip" {
  type    = string
  default = "10.0.1.10"
}
variable "database_private_ip" {
  type    = string
  default = "10.0.2.10"
}
variable "alarm_email" {
  type        = string
  description = "Email endpoint for OCI Notifications. The recipient must confirm OCI's subscription email."
}
