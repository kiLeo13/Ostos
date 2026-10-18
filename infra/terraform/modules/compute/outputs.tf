output "instance_id" { value = oci_core_instance.this.id }
output "public_ip" { value = oci_core_public_ip.this.ip_address }
output "public_ip_id" { value = oci_core_public_ip.this.id }
output "private_ip" { value = oci_core_instance.this.private_ip }

