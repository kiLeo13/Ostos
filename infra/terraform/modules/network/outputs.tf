output "vcn_id" { value = oci_core_vcn.this.id }
output "public_subnet_id" { value = oci_core_subnet.public.id }
output "database_subnet_id" { value = oci_core_subnet.database.id }
output "compute_nsg_id" { value = oci_core_network_security_group.compute.id }
output "mysql_client_nsg_id" { value = oci_core_network_security_group.mysql_client.id }
output "mysql_nsg_id" { value = oci_core_network_security_group.mysql.id }

