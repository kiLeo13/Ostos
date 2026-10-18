variable "compartment_id" { type = string }
variable "name_prefix" { type = string }
variable "vcn_cidr" { type = string }
variable "public_subnet_cidr" { type = string }
variable "database_subnet_cidr" { type = string }
variable "admin_ipv4_cidr" { type = string }
variable "cloudflare_ipv4_cidrs" { type = set(string) }

