resource "oci_mysql_mysql_db_system" "this" {
  availability_domain = var.availability_domain
  compartment_id      = var.compartment_id
  display_name        = var.display_name
  shape_name          = var.shape_name
  subnet_id           = var.subnet_id
  nsg_ids             = var.nsg_ids
  ip_address          = var.private_ip
  admin_username      = var.admin_username
  admin_password      = var.admin_password
  is_highly_available = false

  backup_policy {
    is_enabled        = true
    retention_in_days = 1
  }

  deletion_policy {
    automatic_backup_retention = "RETAIN"
    final_backup               = "SKIP_FINAL_BACKUP"
    is_delete_protected        = true
  }

  secure_connections {
    certificate_generation_type = "SYSTEM"
  }

  freeform_tags = {
    application = "ostos"
  }

  lifecycle {
    prevent_destroy = true
    precondition {
      condition     = var.shape_name == "MySQL.Free"
      error_message = "Ostos production permits only the Always Free MySQL.Free DB system shape."
    }
  }
}

