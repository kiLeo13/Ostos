output "db_system_id" { value = oci_mysql_mysql_db_system.this.id }
output "private_ip" { value = oci_mysql_mysql_db_system.this.ip_address }
output "port" { value = oci_mysql_mysql_db_system.this.port }
output "hostname" { value = try(oci_mysql_mysql_db_system.this.endpoints[0].hostname, null) }
