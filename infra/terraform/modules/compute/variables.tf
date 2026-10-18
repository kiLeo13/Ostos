variable "availability_domain" { type = string }
variable "compartment_id" { type = string }
variable "display_name" { type = string }
variable "image_id" { type = string }
variable "shape" { type = string }
variable "subnet_id" { type = string }
variable "nsg_ids" { type = list(string) }
variable "private_ip" { type = string }
variable "ssh_public_key" { type = string }
variable "snowflake_node_id" { type = number }

