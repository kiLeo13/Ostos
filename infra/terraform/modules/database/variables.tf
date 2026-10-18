variable "availability_domain" { type = string }
variable "compartment_id" { type = string }
variable "display_name" { type = string }
variable "shape_name" { type = string }
variable "subnet_id" { type = string }
variable "nsg_ids" { type = list(string) }
variable "private_ip" { type = string }
variable "admin_username" { type = string }
variable "admin_password" {
  type      = string
  sensitive = true
}
